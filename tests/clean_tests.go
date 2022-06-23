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

func testCleaner() {
	filepath := "tests/cleaner_files/clean_in.asc" // passed through
	// toleranceIsland := 40000.0 // standard tolerance
	// toleranceVoid := 22500.0   // standard tolerance
	// useChunk := false
	toleranceIsland := 4.0 // test smaller datasets
	toleranceVoid := 2.0   // test smaller datasets
	chunkx := 2
	chunky := 3

	filepathTif := tools.ChangeExtension(filepath, ".tif")
	processing.TransferType(filepath, filepathTif, "Byte")
	fmt.Printf("____________________________\nCleaner\n")
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

		correct := fmt.Sprintf("tests/cleaner_files/clean_i%.0fv%.0fd%v_correct.asc", toleranceIsland, toleranceVoid, adjType)
		fmt.Printf("\tNO CHUNKING: %s\t%v\n", outfileFullAsc, sameFiles(outfileFullAsc, correct))
		fmt.Printf("\tCHUNKING: %s\t%v\n", outfileChunkAsc, sameFiles(outfileChunkAsc, correct))

		if !sameFiles(outfileFullAsc, correct) {
			bunyan.Errorf("FILE: %s\t\tincorrect\t| Correct: %s", outfileFullAsc, correct)
		}
		if !sameFiles(outfileChunkAsc, correct) {
			bunyan.Errorf("FILE: %s\tincorrect\t| Correct: %s", outfileChunkAsc, correct)
		}

		//Delete unnecessary files
		for _, fp := range [2]string{outfileFull, outfileChunk} {
			for _, ext := range [3]string{".asc.aux.xml", ".prj", ".tiff"} {
				os.Remove(fp + ext)
			}
		}

	}
	// Delete tif creation
	os.Remove(filepathTif)
	fmt.Printf("____________________________\n")
}
