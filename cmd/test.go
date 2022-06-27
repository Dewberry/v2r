package cmd

import (
	"app/tests"

	bunyan "github.com/Dewberry/paul-bunyan"
	"github.com/spf13/cobra"
)

// cleanCmd represents the clean command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run test suite",
	Long:  `Testing suite for idw and cleaner. Failed tests show at logger.ERROR level`,
	Run: func(cmd *cobra.Command, args []string) {
		bunyan.Info("Test Suite Started")
		tests.TestSuite()
		bunyan.Info("Test Suite Completed")
	},
}

func init() {
	rootCmd.AddCommand(testCmd)

	initLogging()

}
