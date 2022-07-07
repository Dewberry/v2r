package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/dewberry/v2r/features/cleaner"
	"github.com/dewberry/v2r/tools"

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

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Run cleaning algorithm",
	Long:  `Clean islands/voids on a map if they are below the given threshold.`,
	Run: func(cmd *cobra.Command, args []string) {
		bunyan.Info("Cleaner Started")
		printFlagsCleaner()
		clean()
		bunyan.Info("Cleaner Finished")
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)

	cleanCmd.Flags().StringVarP(&filepath, "file", "f", "", "file to run (required)")

	cleanCmd.Flags().BoolVarP(&useChunk, "concurrent", "c", false, "run concurrently or serially")
	cleanCmd.Flags().IntVarP(&adjType, "adjacent", "a", 8, "set adjacency type [4: only cardinal directions | 8: include ordinal]")

	cleanCmd.Flags().Float64Var(&toleranceIsland, "ti", 40000.0, "set tolerance for islands")
	cleanCmd.Flags().Float64Var(&toleranceVoid, "tv", 22500.0, "set tolerance for voids")
	cleanCmd.Flags().IntVar(&cleanChunkX, "cx", 256*10, "set chunk size in x-direction")
	cleanCmd.Flags().IntVar(&cleanChunkY, "cy", 256*10, "set chunk size in y-direction")

}

func printFlagsCleaner() {
	bunyan.Info("-----Flags-----")

	bunyan.Infof("Filepath: %v", filepath)
	bunyan.Infof("Concurrent: %v", useChunk)
	if useChunk {
		bunyan.Infof("Partition (x-direction): %v", cleanChunkX)
		bunyan.Infof("Partition (y-direction): %v", cleanChunkY)
	}
	bunyan.Infof("Adjacency Type: d%v", adjType)
	bunyan.Infof("Tolerance (Islands): %.1f", toleranceIsland)
	bunyan.Infof("Tolerance (Voids): %.1f", toleranceVoid)
	bunyan.Info("---------------")
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
