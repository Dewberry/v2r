package cmd

import (
	"app/tools"
	"os"

	bunyan "github.com/Dewberry/paul-bunyan"
	"github.com/spf13/cobra"
)

var (
	useLumberjack bool
	lvlDebug      bool
	lvlInfo       bool
	lvlError      bool
)

var rootCmd = &cobra.Command{
	Use:   "v2r",
	Short: "Vector to Raster interpolation routines",
	Long:  `Vector to Raster interpolation routines with anaccompanying testing suite.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {

}

func initLogging() {
	rootCmd.PersistentFlags().BoolVarP(&useLumberjack, "log", "l", false, "Log outputs")
	rootCmd.PersistentFlags().BoolVarP(&lvlDebug, "debug", "d", false, "Set debug level to DEBUG")
	rootCmd.PersistentFlags().BoolVarP(&lvlInfo, "info", "i", true, "Set debug level to INFO")
	rootCmd.PersistentFlags().BoolVarP(&lvlError, "error", "e", false, "Set debug level to Error")

	if useLumberjack {
		tools.SetLogging()
	}

	if lvlDebug {
		bunyan.New().SetLevel(bunyan.DEBUG)
	} else if lvlInfo {
		bunyan.New().SetLevel(bunyan.INFO)
	} else if lvlError {
		bunyan.New().SetLevel(bunyan.ERROR)
	}
}
