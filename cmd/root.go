package cmd

import (
	"github.com/dewberry/v2r/tools"

	bunyan "github.com/Dewberry/paul-bunyan"
	"github.com/spf13/cobra"
)

var (
	useLumberjack bool
	lvlDebug      bool
	lvlError      bool
)

var rootCmd = &cobra.Command{
	Use:   "v2r",
	Short: "Vector to Raster interpolation routines",
	Long:  `Vector to Raster interpolation routines with an accompanying testing suite.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		bunyan.Fatal(err)
	}
}

func init() {
	cobra.OnInitialize(initLogging)
	rootCmd.PersistentFlags().BoolVarP(&useLumberjack, "log", "l", false, "Log outputs")
	rootCmd.PersistentFlags().BoolVarP(&lvlDebug, "debug", "d", false, "Set logging level to DEBUG")
	rootCmd.PersistentFlags().BoolVarP(&lvlError, "error", "e", false, "Set logging level to ERROR")
}

func initLogging() {
	if useLumberjack {
		tools.SetLogging()
	}
	if lvlDebug {
		bunyan.New().SetLevel(bunyan.DEBUG)
	} else if lvlError {
		bunyan.New().SetLevel(bunyan.ERROR)
	}
}
