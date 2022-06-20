package processing

import (
	"app/tools"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/dewberry/gdal"
)

type GDalInfo struct {
	XMin         float64
	YMin         float64
	XCell        float64
	YCell        float64
	GDalDataType gdal.DataType
	EPSG         int
}

func CreateGDalInfo(XMin float64, YMin float64, XCell float64, YCell float64, GDalDataType gdal.DataType, EPSG int) GDalInfo {
	return GDalInfo{XMin, YMin, XCell, YCell, GDalDataType, EPSG}
}

func WriteTif(unwrappedMatrix interface{}, GDINFO GDalInfo, filename string, offsets tools.OrderedPair, totalSize tools.OrderedPair, bufferSize tools.OrderedPair, create bool) error {
	filename = fmt.Sprintf("%s.tiff", filename)

	var dataset gdal.Dataset
	if create {
		fmt.Println("Creating Raster")
		driver, err := gdal.GetDriverByName("GTIFF")
		if err != nil {
			log.Fatal(err)
			return err
		}
		// fmt.Printf("Creating dataset\n")
		dataset = driver.Create(filename, totalSize.C, totalSize.R, 1, GDINFO.GDalDataType, nil)

		defer dataset.Close()

		spatialRef := gdal.CreateSpatialReference("")
		spatialRef.FromEPSG(GDINFO.EPSG)
		srString, err := spatialRef.ToWKT()
		if err != nil {
			log.Fatal(err)
			return err
		}
		dataset.SetProjection(srString)

	} else {
		fmt.Println("Updating Raster")
		var err error
		dataset, err = gdal.Open(filename, gdal.Update)
		if err != nil {
			log.Fatal(err)
			return err
		}

		defer dataset.Close()

	}

	dataset.SetGeoTransform([6]float64{GDINFO.XMin, GDINFO.XCell, 0, GDINFO.YMin, 0, GDINFO.YCell})

	raster := dataset.RasterBand(1)
	return raster.IO(gdal.Write, offsets.C, offsets.R, bufferSize.C, bufferSize.R, unwrappedMatrix, bufferSize.C, bufferSize.R, 0, 0)
}

func getEPSG(s string) int {
	loc := strings.LastIndex(s, "ID[\"EPSG\",") + 10
	if loc == -1 {
		return -1
	}
	epsg5, err := strconv.Atoi(s[loc : loc+5])
	if err != nil {
		epsg4, err := strconv.Atoi(s[loc : loc+4])
		if err != nil {
			return -2
		}
		return epsg4
	}
	return epsg5

}

func ReadTif(filepath string, offsets tools.OrderedPair, size tools.OrderedPair, entireFile bool) ([]byte, GDalInfo, tools.OrderedPair, error) {
	DS, err := gdal.Open(filepath, gdal.ReadOnly)
	if err != nil {
		return []byte{}, GDalInfo{}, tools.OrderedPair{}, err
	}
	if entireFile {
		size = tools.MakePair(DS.RasterYSize(), DS.RasterXSize())
	}

	info := gdal.Info(DS, nil)
	inGT := DS.GeoTransform()

	gdReturn := GDalInfo{inGT[0], inGT[3], inGT[1], inGT[5], gdal.DataType(gdal.Byte), getEPSG(info)}

	band := DS.RasterBand(1)

	data := make([]byte, size.C*size.R)
	err = band.IO(gdal.Read, offsets.C, offsets.R, size.C, size.R, data, size.C, size.R, 0, 0)
	if err != nil {
		return []byte{}, gdReturn, size, err
	}

	return data, gdReturn, size, nil
}

func GetTifInfo(filepath string) (GDalInfo, tools.OrderedPair, error) {
	DS, err := gdal.Open(filepath, gdal.ReadOnly)
	if err != nil {
		return GDalInfo{}, tools.OrderedPair{}, err
	}

	numCols := DS.RasterXSize()
	numRows := DS.RasterYSize()

	info := gdal.Info(DS, nil)
	inGT := DS.GeoTransform()

	gdReturn := GDalInfo{inGT[0], inGT[3], inGT[1], inGT[5], gdal.DataType(gdal.Byte), getEPSG(info)}

	return gdReturn, tools.MakePair(numRows, numCols), nil
}
