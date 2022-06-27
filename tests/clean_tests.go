package tests

import (
	"fmt"
	"os"
	"strings"

	cleaner "app/features/cleaner"
	tools "app/tools"
	processing "app/tools/processing"

	bunyan "github.com/Dewberry/paul-bunyan"
)

func testCleaner() bool {
	filepath := "tests/cleaner_files/clean_in.asc" // passed through
	// toleranceIsland := 40000.0 // standard tolerance
	// toleranceVoid := 22500.0   // standard tolerance
	// useChunk := false
	toleranceVoid := 2.0 // test smaller datasets
	chunkx := 2
	chunky := 3

	filepathTif := tools.ChangeExtension(filepath, ".tif")
	processing.TransferType(filepath, filepathTif, "Byte")
	bunyan.Info("____________________________")
	bunyan.Info("Cleaner")
	pass := true
	for _, toleranceIsland := range [2]float64{4.0, 9.0} {
		for _, adjType := range [2]int{4, 8} {

			outfileFull := fmt.Sprintf("%s_isl%.0fvoid%.0f_cleaned%v", strings.TrimSuffix(filepath, ".asc"), toleranceIsland, toleranceVoid, adjType)
			outfileChunk := outfileFull + "chunked"

			err := cleaner.CleanFull(filepathTif, outfileFull, toleranceIsland, toleranceVoid, adjType)
			if err != nil {
				bunyan.Fatal(err)
			}
			err = cleaner.CleanWithChunking(filepathTif, outfileChunk, toleranceIsland, toleranceVoid, tools.MakePair(chunky, chunkx), adjType)
			if err != nil {
				bunyan.Fatal(err)
			}

			outfileFullTif := fmt.Sprintf("%s.tiff", outfileFull)
			outfileFullAsc := fmt.Sprintf("%s.asc", outfileFull)
			processing.TransferType(outfileFullTif, outfileFullAsc, "Int16")

			outfileChunkTif := fmt.Sprintf("%s.tiff", outfileChunk)
			outfileChunkAsc := fmt.Sprintf("%s.asc", outfileChunk)
			processing.TransferType(outfileChunkTif, outfileChunkAsc, "Int16")

			correctFP := fmt.Sprintf("tests/cleaner_files/clean_i%.0fv%.0fd%v_correct", toleranceIsland, toleranceVoid, adjType)
			correct := correctFP + ".asc"
			bunyan.Infof("     NO CHUNKING: %s         %v", outfileFullAsc, sameFiles(outfileFullAsc, correct))
			bunyan.Infof("     CHUNKING: %s     %v", outfileChunkAsc, sameFiles(outfileChunkAsc, correct))

			if !sameFiles(outfileFullAsc, correct) {
				bunyan.Errorf("FILE: %s  |  Correct: %s", outfileFullAsc, correct)
				pass = false
			}
			if !sameFiles(outfileChunkAsc, correct) {
				bunyan.Errorf("FILE: %s  |  Correct: %s", outfileChunkAsc, correct)
				pass = false
			}

			//Delete unnecessary files
			for _, fp := range [3]string{outfileFull, outfileChunk, correctFP} {
				for _, ext := range [3]string{".asc.aux.xml", ".prj", ".tiff"} {
					os.Remove(fp + ext)
				}
			}
		}
	}
	// Delete tif creation
	os.Remove(filepathTif)
	bunyan.Info("____________________________")
	return pass
}
