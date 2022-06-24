package processing

import (
	"app/tools"
	"fmt"
	"strconv"
	"strings"

	bunyan "github.com/Dewberry/paul-bunyan"
	"github.com/dewberry/gdal"
)

type GDalInfo struct {
	XMin         float64
	YMin         float64
	XCell        float64
	YCell        float64
	GDalDataType gdal.DataType
	Proj         string
}

func CreateGDalInfo(XMin float64, YMin float64, XCell float64, YCell float64, GDalDataType gdal.DataType, proj string) GDalInfo {
	return GDalInfo{XMin, YMin, XCell, YCell, GDalDataType, proj}
}

func WriteTif(unwrappedMatrix interface{}, GDINFO GDalInfo, filename string, offsets tools.OrderedPair, totalSize tools.OrderedPair, bufferSize tools.OrderedPair, create bool) error {
	filename = fmt.Sprintf("%s.tiff", filename)
	return WriteGDAL(unwrappedMatrix, GDINFO, filename, "GTiff", offsets, totalSize, bufferSize, create)
}

func WriteAscii(unwrappedMatrix interface{}, GDINFO GDalInfo, filename string, offsets tools.OrderedPair, totalSize tools.OrderedPair, bufferSize tools.OrderedPair, create bool) error {
	filename = fmt.Sprintf("%s.asc", filename)
	return WriteGDAL(unwrappedMatrix, GDINFO, filename, "ASCIIGRID", offsets, totalSize, bufferSize, create)
}

func WriteGDAL(unwrappedMatrix interface{}, GDINFO GDalInfo, filename string, driver string, offsets tools.OrderedPair, totalSize tools.OrderedPair, bufferSize tools.OrderedPair, create bool) error {
	var dataset gdal.Dataset
	if create {
		// fmt.Println("Creating Raster")
		driver, err := gdal.GetDriverByName(driver)
		if err != nil {
			bunyan.Fatal(err)
			return err
		}
		// fmt.Printf("Creating dataset\n")
		dataset = driver.Create(filename, totalSize.C, totalSize.R, 1, GDINFO.GDalDataType, []string{"BIGTIFF=YES"})

		defer dataset.Close()

		dataset.SetProjection(GDINFO.Proj)
		dataset.SetGeoTransform([6]float64{GDINFO.XMin, GDINFO.XCell, 0, GDINFO.YMin, 0, GDINFO.YCell})

	} else {
		// fmt.Println("Updating Raster")
		var err error
		dataset, err = gdal.Open(filename, gdal.Update)
		if err != nil {
			bunyan.Fatal(err)
			return err
		}

		defer dataset.Close()

	}

	raster := dataset.RasterBand(1)
	return raster.IO(gdal.Write, offsets.C, offsets.R, bufferSize.C, bufferSize.R, unwrappedMatrix, bufferSize.C, bufferSize.R, 0, 0)
}

func getEPSG(s string) int {
	loc := strings.LastIndex(s, "ID[\"EPSG\",") + 10
	if loc == -1 {
		bunyan.Info("invalid or no epsg given, 2284 used instead")
		return 2284
	}
	epsg5, err := strconv.Atoi(s[loc : loc+5])
	if err != nil {
		epsg4, err := strconv.Atoi(s[loc : loc+4])
		if err != nil {
			bunyan.Info("invalid or no epsg given, 2284 used instead")
			return 2284
		}
		return epsg4
	}
	return epsg5

}

func ReadGDAL(filepath string, offsets tools.OrderedPair, size tools.OrderedPair, entireFile bool) ([]byte, GDalInfo, tools.OrderedPair, error) {
	DS, err := gdal.Open(filepath, gdal.ReadOnly)
	if err != nil {
		return []byte{}, GDalInfo{}, tools.OrderedPair{}, err
	}
	if entireFile {
		size = tools.MakePair(DS.RasterYSize(), DS.RasterXSize())
	}

	inGT := DS.GeoTransform()

	gdReturn := GDalInfo{inGT[0], inGT[3], inGT[1], inGT[5], gdal.DataType(gdal.Byte), DS.Projection()}

	band := DS.RasterBand(1)

	data := make([]byte, size.C*size.R)
	err = band.IO(gdal.Read, offsets.C, offsets.R, size.C, size.R, data, size.C, size.R, 0, 0)
	if err != nil {
		return []byte{}, gdReturn, size, err
	}

	return data, gdReturn, size, nil
}

func GetInfoGDAL(filepath string) (GDalInfo, tools.OrderedPair, error) {
	DS, err := gdal.Open(filepath, gdal.ReadOnly)
	if err != nil {
		return GDalInfo{}, tools.OrderedPair{}, err
	}

	numCols := DS.RasterXSize()
	numRows := DS.RasterYSize()

	inGT := DS.GeoTransform()
	// bunyan.Debug(info)
	gdReturn := GDalInfo{inGT[0], inGT[3], inGT[1], inGT[5], gdal.DataType(gdal.Byte), DS.Projection()}

	return gdReturn, tools.MakePair(numRows, numCols), nil
}

func TransferType(src string, dst string, outputType string) {
	DS, err := gdal.Open(src, gdal.ReadOnly)
	if err != nil {
		bunyan.Fatal(err)
	}

	// datatype := gdal.Int16
	opts := []string{"-ot", outputType}
	if strings.HasSuffix(dst, ".tiff") || strings.HasSuffix(dst, ".tif") {
		opts = append(opts, "-of", "GTiff",
			"-co", "TILED=YES",
			"-co", "COPY_SRC_OVERVIEWS=YES")
	}
	_, err = gdal.Translate(dst, DS, opts)
	if err != nil {
		bunyan.Fatal(err)
	}

	// gdal.GDALTranslateOptions()
}
