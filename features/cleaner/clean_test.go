package cleaner

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/dewberry/v2r/tools"
	"github.com/dewberry/v2r/tools/processing"

	bunyan "github.com/Dewberry/paul-bunyan"
)

func TestCleaner(t *testing.T) {
	bunyan.Info("____________________________")
	bunyan.Info("Cleaner")

	for _, cs := range [2]string{"Serial", "Conc"} {
		for _, toleranceIsland := range [2]float64{4.0, 9.0} {
			for _, adjType := range [2]int{4, 8} {
				t.Run(fmt.Sprintf("%s_T%v_A%v", cs, toleranceIsland, adjType), func(t *testing.T) {
					if !cleanerTestHelper(toleranceIsland, adjType, cs == "Conc") {
						t.Error()
					}
				})
			}
		}
	}

	bunyan.Info("____________________________")
}

func cleanerTestHelper(toleranceIsland float64, adjType int, chunk bool) bool {
	filepath := "cleaner_test_files/clean_in.asc"
	toleranceVoid := 2.0 // test smaller datasets
	chunkx := 2
	chunky := 3

	chunkString := "   "
	outfile := fmt.Sprintf("%s_isl%.0fvoid%.0f_cleaned%v", strings.TrimSuffix(filepath, ".asc"), toleranceIsland, toleranceVoid, adjType)
	correctFP := fmt.Sprintf("cleaner_test_files/clean_i%.0fv%.0fd%v_correct", toleranceIsland, toleranceVoid, adjType)

	if chunk {
		outfile += "chunked"
		err := CleanWithChunking(filepath, outfile, toleranceIsland, toleranceVoid, tools.MakePair(chunky, chunkx), adjType)
		if err != nil {
			bunyan.Error("CleanWithChunking() error")
			return false
		}
	} else {
		chunkString = "NO "
		err := CleanFull(filepath, outfile, toleranceIsland, toleranceVoid, adjType)
		if err != nil {
			bunyan.Error("CleanFull() error")
			return false
		}
	}

	processing.TransferType(outfile+".tiff", outfile+".asc", "Int16")

	//Delete unnecessary files
	for _, fp := range [2]string{outfile, correctFP} {
		for _, ext := range [3]string{".asc.aux.xml", ".prj", ".tiff"} {
			os.Remove(fp + ext)
		}
	}

	//File Comparison
	isCorrect, err := tools.SameFiles(outfile+".asc", correctFP+".asc")
	if err != nil {
		bunyan.Fatal(err)
	}
	bunyan.Infof("     %sCHUNKING: %s     %v", chunkString, outfile+".asc", isCorrect)
	if !isCorrect {
		bunyan.Errorf("FILE: %s  |  Correct: %s", outfile+".asc", correctFP+".asc")
	}
	return isCorrect
}
