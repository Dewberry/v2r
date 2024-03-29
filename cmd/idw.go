package cmd

import (
	"fmt"
	"strings"
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

	cobra.OnInitialize(reqFlagsIDW, expRange)

	idwCmd.Flags().BoolVarP(&fromGPKG, "gpkg", "g", false, "read from gpkg (true) or from txt file (false)")
	idwCmd.Flags().BoolVarP(&useChunking, "concurrent", "c", false, "run program concurrently (true) or serially (false)")

	idwCmd.Flags().IntVar(&idwChunkX, "cx", 200, "set chunk size in x-direction")
	idwCmd.Flags().IntVar(&idwChunkY, "cy", 200, "set chunk size in y-direction")
	idwCmd.Flags().IntVar(&epsg, "epsg", 2284, "set EPSG code")

	idwCmd.Flags().Float64Var(&expIncrement, "ei", .5, "set exponential incremement for calculations between start and end")
	idwCmd.Flags().Float64Var(&expStart, "es", 1.5, "set start for exponent (inclusive)")
	idwCmd.Flags().Float64Var(&expEnd, "ee", 0.0, "set end for exponent (exclusive)")
	idwCmd.Flags().Float64Var(&stepX, "sx", 100.0, "set step size in x-direction")
	idwCmd.Flags().Float64Var(&stepY, "sy", 100.0, "set step size in y-direction")

	idwCmd.Flags().StringVarP(&infile, "file", "f", "", "set filepath (required)")
	idwCmd.Flags().StringVar(&outfileFolder, "outPath", "data/idw/", "set outfile location")

	idwCmd.Flags().StringVar(&layer, "layer", "", "set name of layer in geopackage file (*)")
	idwCmd.Flags().StringVar(&field, "field", "", "set name of field in geopackage file (*)")

}

func reqFlagsIDW() {
	if fromGPKG {
		idwCmd.MarkFlagRequired("layer")
		idwCmd.MarkFlagRequired("field")
	}
}

func expRange() {
	bunyan.Info("exp")
	if !idwCmd.Flag("ee").Changed { // default usage
		expEnd = expStart + expIncrement
	}
	if expIncrement <= 0 {
		bunyan.Fatalf("exponential increment (%v) must be > 0", expIncrement)
	}
	if expEnd <= expStart {
		bunyan.Fatalf("exponent range: [%v, %v) has no iterations", expStart, expEnd)
	}

}

func printFlagsIDW() {
	bunyan.Info("-----Flags-----")

	bunyan.Infof("Filepath: %v", infile)
	bunyan.Infof("Outfile folder: %v", outfileFolder)
	bunyan.Infof("Concurrent: %v", useChunking)
	if useChunk {
		bunyan.Infof("Partition (x-direction): %v", idwChunkX)
		bunyan.Infof("Partition (y-direction): %v", idwChunkY)
	}
	bunyan.Infof("From GPKG: %v", fromGPKG)
	if fromGPKG {
		bunyan.Infof("Layer name: %v", layer)
		bunyan.Infof("Field name: %v", field)
	} else {
		bunyan.Infof("Used epsg = %v on text file", epsg)
	}
	bunyan.Infof("Step size in x-direction: %v", stepX)
	bunyan.Infof("Step size in y-direction: %v", stepY)

	bunyan.Infof("Exponent: [%v, %v)   Step size: %v", expStart, expEnd, expIncrement)
	bunyan.Info("---------------")

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

	outfilePrefix := infile[tools.Max(strings.LastIndex(infile, "/")+1, strings.LastIndex(infile, "\\")+1):strings.LastIndex(infile, ".")]
	outfile := fmt.Sprintf("%s%s_step%.0f-%.0f%s", outfileFolder, outfilePrefix, stepX, stepY, chunkString) // "{prefix}_step{x}-{y}[chunked]exp{exp}.[ext]"
	iterations := 0
	channel := make(chan string, iterations)

	bunyan.Debug(outfile)
	for exp := expStart; exp < expEnd; exp += expIncrement {
		iterations++
		outfile_formatted := fmt.Sprintf("%spow%.1f.tiff", outfile, exp)
		if !useChunking {
			go idw.FullSolve(&data, outfile_formatted, xInfo, yInfo, proj, exp, channel)
		} else {
			go idw.ChunkSolve(&data, outfile_formatted, xInfo, yInfo, idwChunkY, idwChunkX, proj, exp, channel)
		}
	}

	for exp := expStart; exp < expEnd; exp += expIncrement {
		receivedString := <-channel
		bunyan.Debug(receivedString)
	}

	bunyan.Infof("Completed %v iterations in %v", iterations, time.Since(start))
	bunyan.Infof("Outfiles: %vpow{EXP}.{EXT}", outfile)
}
