package processing

import (
	"fmt"
	"strings"

	"github.com/dewberry/v2r/tools"

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

// Write out cleaned data using gdal. Currently only supports ascii and tiff types.
func WriteGDAL(unwrappedMatrix interface{}, GDINFO GDalInfo, filename string, offsets tools.OrderedPair, totalSize tools.OrderedPair, bufferSize tools.OrderedPair, create bool) error {
	var dataset gdal.Dataset
	if create {
		bunyan.Debug("Creating Raster")

		ext_driver := map[string]string{
			".tif":  "GTiff",
			".tiff": "GTiff",
			".asc":  "ASCIIGRID",
		}
		driverShortName := ""

		for ext, _ := range ext_driver {
			if strings.HasSuffix(filename, ext) {
				driverShortName = ext_driver[ext]
			}
		}

		if driverShortName == "" {
			return fmt.Errorf("Not a valid output file extension. Only accepts tif and ascii files (.tif/.tiff and .asc).")
		}

		driver, err := gdal.GetDriverByName(driverShortName)
		if err != nil {
			return err
		}
		bunyan.Debug("Creating Dataset")
		dataset = driver.Create(filename, totalSize.C, totalSize.R, 1, GDINFO.GDalDataType, []string{"BIGTIFF=YES"})

		defer dataset.Close()

		dataset.SetProjection(GDINFO.Proj)
		dataset.SetGeoTransform([6]float64{GDINFO.XMin, GDINFO.XCell, 0, GDINFO.YMin, 0, GDINFO.YCell})

	} else {
		var err error
		dataset, err = gdal.Open(filename, gdal.Update)
		if err != nil {
			return err
		}

		defer dataset.Close()

	}

	raster := dataset.RasterBand(1)
	return raster.IO(gdal.Write, offsets.C, offsets.R, bufferSize.C, bufferSize.R, unwrappedMatrix, bufferSize.C, bufferSize.R, 0, 0)
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
		return []byte{}, GDalInfo{}, tools.OrderedPair{}, err
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
	gdReturn := GDalInfo{inGT[0], inGT[3], inGT[1], inGT[5], gdal.DataType(gdal.Byte), DS.Projection()}

	return gdReturn, tools.MakePair(numRows, numCols), nil
}

func TransferType(src string, dst string, outputType string) error {
	DS, err := gdal.Open(src, gdal.ReadOnly)
	if err != nil {
		return err
	}

	opts := []string{"-ot", outputType}
	if strings.HasSuffix(dst, ".tiff") || strings.HasSuffix(dst, ".tif") {
		opts = append(opts, "-of", "GTiff",
			"-co", "TILED=YES",
			"-co", "COPY_SRC_OVERVIEWS=YES")
	}
	_, err = gdal.Translate(dst, DS, opts)
	if err != nil {
		return err
	}
	return nil
}

func validGPKGLayer(filepath string, layer string) error {
	ds := gdal.OpenDataSource(filepath, int(gdal.ReadOnly))
	defer ds.Destroy()

	allLayers := make([]string, 0, ds.LayerCount())
	for i := 0; i < ds.LayerCount(); i++ {
		if ds.LayerByIndex(i).Name() == layer {
			return nil
		}
		allLayers = append(allLayers, ds.LayerByIndex(i).Name())
	}
	return fmt.Errorf("Invalid Layer: %v  | %v Possible Layers: %v", layer, len(allLayers), allLayers)
}

func getFieldIndex(fieldDef gdal.FeatureDefinition, field string) (int, error) {
	ind := fieldDef.FieldIndex(field)
	if ind >= 0 {
		return ind, nil
	}
	allFields := make([]string, 0, fieldDef.FieldCount())
	for i := 0; i < fieldDef.FieldCount(); i++ {
		allFields = append(allFields, fieldDef.FieldDefinition(i).Name())
	}

	return -1, fmt.Errorf("Invalid Field: %v  | %v Possible Fields: %v", field, len(allFields), allFields)
}

func getGPKGPoints(filepath string, layer string, field string) ([]tools.Point, string, error) {
	bunyan.Debugf("%s %s %s", filepath, layer, field)
	err := validGPKGLayer(filepath, layer)
	if err != nil {
		return []tools.Point{}, "", err
	}

	ds := gdal.OpenDataSource(filepath, int(gdal.ReadOnly))
	defer ds.Destroy()

	l := ds.LayerByName(layer)
	bunyan.Debug((l.Name()))

	fieldDef := l.Definition()

	fieldIndex, err := getFieldIndex(fieldDef, field)
	if err != nil {
		return []tools.Point{}, "", err
	}
	bunyan.Debug("field index", fieldIndex)

	count, _ := l.FeatureCount(true)

	pointList := make([]tools.Point, 0, fieldDef.FieldCount())
	for i := 1; i < count+1; i++ {
		feature := l.Feature(int64(i))
		geom := feature.Geometry()
		pointList = append(pointList, tools.MakePoint(geom.X(0), geom.Y(0), feature.FieldAsFloat64(fieldIndex)))
	}

	proj, err := l.SpatialReference().ToWKT()
	if err != nil {
		return []tools.Point{}, "", err
	}

	return pointList, proj, nil
}
