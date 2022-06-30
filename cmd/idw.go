package cmd

import (
	"fmt"
	"time"

	"github.com/dewberry/v2r/features/idw"
	"github.com/dewberry/v2r/tools"
	"github.com/dewberry/v2r/tools/processing"

	bunyan "github.com/Dewberry/paul-bunyan"
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

var idwCmd = &cobra.Command{
	Use:   "idw",
	Short: "Run Inverse Distance Weighting algorithm ",
	Long:  `Run Inverse Distance Weighting for multivariate interpolation given a set of points that contain (x, y, elevation).  `,
	Run: func(cmd *cobra.Command, args []string) {
		bunyan.Info("IDW Started")
		printFlagsIDW()
		doIDW()
		bunyan.Info("IDW Finished")
	},
}

func init() {
	rootCmd.AddCommand(idwCmd)

	cobra.OnInitialize(reqFlagsIDW)

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

	idwCmd.Flags().StringVar(&layer, "layer", "", "Set name of layer in geopackage file (*)")
	idwCmd.Flags().StringVar(&field, "field", "", "Set name of field in geopackage file (*)")

}

func reqFlagsIDW() {
	if fromGPKG {
		bunyan.Debug("here)")
		idwCmd.MarkFlagRequired("layer")
		idwCmd.MarkFlagRequired("field")
	}
}

func printFlagsIDW() {
	bunyan.Debug("Flags")
	bunyan.Debug("-----")

	bunyan.Debugf("Filepath: %v", infile)
	bunyan.Debugf("Outfile folder: %v", outfileFolder)
	bunyan.Debugf("Concurrent: %v", useChunking)
	if useChunk {
		bunyan.Debugf("Partition (x-direction): %v", idwChunkX)
		bunyan.Debugf("Partition (y-direction): %v", idwChunkY)
	} else {
		bunyan.Debugf("Print to ascii: %v", outAscii)
		bunyan.Debugf("Print to excel: %v", outExcel)
	}
	bunyan.Debugf("From GPKG: %v", fromGPKG)
	if fromGPKG {
		bunyan.Debugf("Layer name: %v", layer)
		bunyan.Debugf("Field name: %v", field)
	} else {
		bunyan.Debugf("used epsg = %v on text file", epsg)
	}
	bunyan.Debugf("Step size in x-direction: %v", stepX)
	bunyan.Debugf("Step size in y-direction: %v", stepY)
	bunyan.Debugf("Exponent: [%v, %v]   Step size: %v", expStart, expEnd, expIncrement)
	bunyan.Debug("-----")

}

func doIDW() {
	start := time.Now()

	var (
		listPoints []tools.Point
		xInfo      tools.Info
		yInfo      tools.Info
		proj       string
		err        error
	)

	if fromGPKG {
		listPoints, proj, xInfo, yInfo, err = processing.ReadGeoPackage(infile, layer, field, stepX, stepY)
		if err != nil {
			bunyan.Fatal(err)
		}

	} else {
		listPoints, proj, xInfo, yInfo, err = processing.ReadTextData(infile, epsg)
		if err != nil {
			bunyan.Fatal(err)
		}
	}
	bunyan.Debug("projection", proj)

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
