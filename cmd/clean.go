package cmd

import (
	"app/features/cleaner"
	"app/tools"
	"fmt"
	"strings"
	"time"

	bunyan "github.com/Dewberry/paul-bunyan"
	"github.com/spf13/cobra"
)

var (
	useChunk        bool
	filepath        string
	adjType         int
	toleranceIsland float64
	toleranceVoid   float64
	cleanChunkX     int
	cleanChunkY     int
)

// cleanCmd represents the clean command
var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Run cleaning algorithm",
	Long:  `Clean islands/voids on a map if they are below the given threshold.`,
	Run: func(cmd *cobra.Command, args []string) {
		bunyan.Info("Cleaner Started")
		clean()
		bunyan.Info("Cleaner Finished")
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)

	cleanCmd.Flags().StringVarP(&filepath, "file", "f", "", "File to run [Needed]")

	cleanCmd.Flags().BoolVarP(&useChunk, "concurrent", "c", true, "Run concurrently or serially")
	cleanCmd.Flags().IntVarP(&adjType, "adjacent", "a", 8, "Set adjacency type 4 (NSEW only), 8 (diagonals included)")

	cleanCmd.Flags().Float64Var(&toleranceIsland, "ti", 40000.0, "Tolerance for islands")
	cleanCmd.Flags().Float64Var(&toleranceVoid, "tv", 22500.0, "Tolerance for voids")
	cleanCmd.Flags().IntVar(&cleanChunkX, "cx", 256*10, "Chunk size in x-direction")
	cleanCmd.Flags().IntVar(&cleanChunkY, "cy", 256*10, "Chunk size in y-direction")

	initLogging()
}

func clean() {
	start := time.Now()

	chunkString := ""
	if useChunk {
		chunkString = "chunked"
	}
	outfile := fmt.Sprintf("%s_isl%.0fvoid%.0f_cleaned%v%v", strings.TrimSuffix(filepath, ".tif"), toleranceIsland, toleranceVoid, adjType, chunkString)

	err := error(nil)
	if useChunk {
		err = cleaner.CleanWithChunking(filepath, outfile, toleranceIsland, toleranceVoid, tools.MakePair(cleanChunkY, cleanChunkX), adjType)
	} else {
		err = cleaner.CleanFull(filepath, outfile, toleranceIsland, toleranceVoid, adjType)
	}

	if err != nil {
		bunyan.Fatal(err)
	}

	bunyan.Infof("Outfile: %s", outfile)
	bunyan.Infof("Finished cleaning in %v", time.Since(start))

}