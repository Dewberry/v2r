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
	Long: `v2r is a library that implements Vector to Raster interpolation routines. 
Currently capabilities include:
	- IDW (Inverse Distance Weighting)
	- Cleaner (used to clean wet/dry spots from maps)`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		bunyan.Fatal(err)
	}
}

func init() {
	cobra.OnInitialize(initLogging)
	rootCmd.PersistentFlags().BoolVarP(&useLumberjack, "log", "l", false, "set logging level")
	rootCmd.PersistentFlags().BoolVarP(&lvlDebug, "debug", "d", false, "set logging level to DEBUG")
	rootCmd.PersistentFlags().BoolVarP(&lvlError, "error", "e", false, "set logging level to ERROR")
}

func initLogging() {
	if useLumberjack {
		tools.SetLogging()
	}
	if lvlDebug {
		bunyan.New().SetLevel(bunyan.DEBUG)
		bunyan.Debug("running in debug mode")
	} else if lvlError {
		bunyan.New().SetLevel(bunyan.ERROR)
	}
}
