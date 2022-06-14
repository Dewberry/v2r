package tools

import (
	"fmt"

	"github.com/dewberry/gdal"
)

func WriteTif(matrix [][]float64, filename string, pow float64) error {
	filename = fmt.Sprintf("%s.tiff", filename)

	fmt.Printf("Loading driver\n")
	driver, err := gdal.GetDriverByName("GTIFF")
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	fmt.Printf("Creating dataset\n")
	dataset := driver.Create(filename, len(matrix[0]), len(matrix), 1, gdal.Float64, nil)
	defer dataset.Close()

	fmt.Printf("Creating projection\n")
	spatialRef := gdal.CreateSpatialReference("")

	fmt.Printf("Setting EPSG code\n")
	spatialRef.FromEPSG(2284)

	fmt.Printf("Converting to WKT\n")
	srString, err := spatialRef.ToWKT()
	if err != nil {
		return err
	}

	fmt.Printf("Assigning projection: %s\n", srString)
	dataset.SetProjection(srString)

	fmt.Printf("Setting geotransform\n")
	dataset.SetGeoTransform([6]float64{GlobalX[0], CELL, 0, GlobalY[0], 0, CELL})

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
