package tools

import (
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

func CreateGDalInfo(XMin float64, YMin float64, XCell float64, YCell float64, GDalDataType gdal.DataType, ESPG int) GDalInfo {
	return GDalInfo{XMin, YMin, XCell, YCell, GDalDataType, ESPG}
}

func WriteTif(matrix [][]float64, GDINFO GDalInfo, filename string) error {
	filename = fmt.Sprintf("%s.tiff", filename)

	fmt.Printf("Loading driver\n")
	driver, err := gdal.GetDriverByName("GTIFF")
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	fmt.Printf("Creating dataset\n")
	dataset := driver.Create(filename, len(matrix[0]), len(matrix), 1, GDINFO.GDalDataType, nil)
	defer dataset.Close()

	fmt.Printf("Creating projection\n")
	spatialRef := gdal.CreateSpatialReference("")

	fmt.Printf("Setting EPSG code\n")
	spatialRef.FromEPSG(GDINFO.EPSG)

	fmt.Printf("Converting to WKT\n")
	srString, err := spatialRef.ToWKT()
	if err != nil {
		return err
	}

	fmt.Printf("Assigning projection: %s\n", srString)
	dataset.SetProjection(srString)

	fmt.Printf("Setting geotransform\n")
	dataset.SetGeoTransform([6]float64{GDINFO.XMin, GDINFO.XCell, 0, GDINFO.YMin, 0, GDINFO.YCell})

	fmt.Printf("Getting raster band\n")
	raster := dataset.RasterBand(1)

	unwrappedMatrix := make([]float64, len(matrix)*len(matrix[0]))
	for r := 0; r < len(matrix); r++ {
		for c := 0; c < len(matrix[0]); c++ {
			unwrappedMatrix[r*len(matrix[0])+c] = matrix[r][c]
		}
	}
	fmt.Printf("Writing to raster band\n")
	return raster.IO(gdal.Write, 0, 0, len(matrix[0]), len(matrix), unwrappedMatrix, len(matrix[0]), len(matrix), 0, 0)
}

func getESPG(s string) int {
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

func ReadTif(filepath string) ([]byte, GDalInfo, OrderedPair, error) {
	DS, err := gdal.Open(filepath, gdal.ReadOnly)
	if err != nil {
		return []byte{}, GDalInfo{}, OrderedPair{}, err
	}

	numCols := DS.RasterXSize()
	numRows := DS.RasterYSize()
	var rowsAndCols OrderedPair = OrderedPair{numRows, numCols}

	info := gdal.Info(DS, nil)
	inGT := DS.GeoTransform()

	gdReturn := GDalInfo{inGT[0], inGT[3], inGT[1], inGT[5], gdal.DataType(gdal.Byte), getESPG(info)}

	band := DS.RasterBand(1)

	data := make([]byte, numCols*numRows)
	err = band.IO(gdal.Read, 0, 0, numCols, numRows, data, numCols, numRows, 0, 0)
	if err != nil {
		return []byte{}, gdReturn, rowsAndCols, err
	}

	return data, gdReturn, rowsAndCols, nil
}

func ReadTifChunk(filepath string, offsets OrderedPair, size OrderedPair) ([]byte, error) {
	// fmt.Printf("filepath: %s\noffsets: %v\nsize: %v\n", filepath, start, size)

	DS, err := gdal.Open(filepath, gdal.ReadOnly)
	if err != nil {
		return []byte{}, err
	}

	// xSize := DS.RasterXSize()
	// ySize := DS.RasterYSize()

	band := DS.RasterBand(1)

	data := make([]byte, size.C*size.R)
	err = band.IO(gdal.Read, offsets.C, offsets.R, size.C, size.R, data, size.C, size.R, 0, 0)
	if err != nil {
		return []byte{}, err
	}

	return data, nil
}

func WriteTifSquare(unwrappedMatrix []byte, GDINFO GDalInfo, offsets OrderedPair, totalSize OrderedPair, bufferSize OrderedPair, filename string, create bool) error {
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
	fmt.Printf("Writing to raster band\n")
	return raster.IO(gdal.Write, offsets.C, offsets.R, bufferSize.C, bufferSize.R, unwrappedMatrix, bufferSize.C, bufferSize.R, 0, 0)
}

func GetTifInfo(filepath string) (GDalInfo, OrderedPair, error) {
	DS, err := gdal.Open(filepath, gdal.ReadOnly)
	if err != nil {
		return GDalInfo{}, OrderedPair{}, err
	}

	numCols := DS.RasterXSize()
	numRows := DS.RasterYSize()

	info := gdal.Info(DS, nil)
	inGT := DS.GeoTransform()

	gdReturn := GDalInfo{inGT[0], inGT[3], inGT[1], inGT[5], gdal.DataType(gdal.Byte), getESPG(info)}

	return gdReturn, OrderedPair{numRows, numCols}, nil
}
