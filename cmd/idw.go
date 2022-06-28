/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (

	// "app/features/idw"

	"app/features/idw"
	"app/tools"
	"app/tools/processing"
	"fmt"
	"time"

	bunyan "github.com/Dewberry/paul-bunyan"
	"github.com/dewberry/gdal"
	"github.com/spf13/cobra"
)

var (
	fromGPKG      bool
	useChunking   bool
	outExcel      bool
	outAscii      bool
	idwChunkX     int
	idwChunkY     int
	expIncrement  float64
	expStart      float64
	expEnd        float64
	stepX         float64
	stepY         float64
	epsg          int
	infile        string
	layer         string
	field         string
	outfileFolder string
)

// idwCmd represents the idw command
var idwCmd = &cobra.Command{
	Use:   "idw",
	Short: "Run Inverse Distance Weighting algorithm ",
	Long:  `Run Inverse Distance Weighting for multivariate interpolation given a set of points that contain (x, y, elevation).  `,
	Run: func(cmd *cobra.Command, args []string) {
		bunyan.Info("IDW Started")
		doIDW()
		bunyan.Info("IDW Finished")
	},
}

func init() {
	rootCmd.AddCommand(idwCmd)

	idwCmd.Flags().BoolVarP(&fromGPKG, "gpkg", "g", false, "Read from gpkg (true) or from txt file (false)")
	idwCmd.Flags().BoolVarP(&useChunking, "concurrent", "c", false, "Run program concurrently (true) or serially (false)")
	idwCmd.Flags().BoolVar(&outAscii, "ascii", false, "Write to ascii file?")
	idwCmd.Flags().BoolVar(&outExcel, "excel", false, "Write to excel spreadsheet?")

	idwCmd.Flags().IntVar(&idwChunkX, "cx", 200, "Set chunk size in x-direction")
	idwCmd.Flags().IntVar(&idwChunkY, "cy", 200, "St chunk size in y-direction")
	idwCmd.Flags().IntVar(&epsg, "epsg", 2284, "Set EPSG code")

	idwCmd.Flags().Float64Var(&expIncrement, "ei", .5, "Exponential incremement for calculations between start and end")
	idwCmd.Flags().Float64Var(&expStart, "es", 1.5, "Start for exponent (inclusive)")
	idwCmd.Flags().Float64Var(&expEnd, "ee", 1.5, "End for exponent (inclusive)")
	idwCmd.Flags().Float64Var(&stepX, "sx", 100.0, "Step size in x-direction")
	idwCmd.Flags().Float64Var(&stepY, "sy", 100.0, "Step size in y-direction")

	idwCmd.Flags().StringVarP(&infile, "file", "f", "", "Set filepath (required)")
	idwCmd.Flags().StringVar(&outfileFolder, "outPath", "data/idw/", "Set outfile location")
	idwCmd.Flags().StringVar(&layer, "layer", "layer1", "Set name of layer in geopackage file")
	idwCmd.Flags().StringVar(&field, "field", "elevation", "Set name of field in geopackage file")
}

func doIDW() {
	start := time.Now()

	var (
		listPoints []tools.Point
		xInfo      tools.Info
		yInfo      tools.Info
		srs        gdal.SpatialReference
		proj       string
		err        error
	)

	if fromGPKG {
		listPoints, srs, xInfo, yInfo = processing.ReadGeoPackage(infile, layer, field, stepX, stepY)

	} else {
		srs = gdal.CreateSpatialReference("")
		err = srs.FromEPSG(epsg)
		if err != nil {
			bunyan.Fatal(err)
		}
		listPoints, xInfo, yInfo, err = processing.ReadData(infile)
		if err != nil {
			bunyan.Fatal(err)
		}
	}

	proj, err = srs.ToWKT()
	if err != nil {
		bunyan.Fatal(err)
	}

	data := tools.MakeCoordSpace(&listPoints, xInfo, yInfo)
	chunkString := ""
	if useChunking {
		chunkString = "chunked"
	}
	outfile := fmt.Sprintf("%sstep%.0f-%.0f%s", outfileFolder, stepX, stepY, chunkString) // "step{x}-{y}[chunked]exp{exp}.[ext]"
	iterations := 1 + int((expEnd-expStart)/expIncrement)
	channel := make(chan string, iterations)

	for exp := expStart; exp <= expEnd; exp += expIncrement {
		if !useChunking {
			go idw.FullSolve(&data, outfile, xInfo, yInfo, proj, exp, outAscii, outExcel, channel)
		} else {
			go idw.ChunkSolve(&data, outfile, xInfo, yInfo, idwChunkY, idwChunkX, proj, exp, channel)
		}
	}

	for exp := expStart; exp <= expEnd; exp += expIncrement {
		receivedString := <-channel
		bunyan.Infof(receivedString)
	}

	bunyan.Infof("Completed %v iterations in %v", iterations, time.Since(start))
	bunyan.Infof("Outfiles: %vpow{EXP}.{EXT}", outfile)
}
