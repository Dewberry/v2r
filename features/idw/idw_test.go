package idw

import (
	"fmt"
	"os"
	"testing"

	"github.com/dewberry/v2r/tools"
	"github.com/dewberry/v2r/tools/processing"

	bunyan "github.com/Dewberry/paul-bunyan"
)

func TestIDW(t *testing.T) {
	bunyan.Info("_____________IDW_____________")
	epsg := 2284

	filepath := "idw_test_files/idw_in.txt"
	listPoints, proj, xInfo, yInfo, err := processing.ReadTextData(filepath, epsg)
	if err != nil {
		t.Error(err)
	}

	data := tools.MakeCoordSpace(&listPoints, xInfo, yInfo)
	for _, cs := range [2]string{"Serial", "Conc"} {
		for _, stepxy := range [2][2]float64{{1.0, 1.0}, {2.0, 2.0}} {
			t.Run(fmt.Sprintf("%v_step%v-%v", cs, stepxy[0], stepxy[1]), func(t *testing.T) {
				if !idwTestHelper(&data, xInfo, yInfo, proj, cs == "Conc") {
					t.Error()
				}
			})
		}
	}
	bunyan.Info("_____________________________")
}

func idwTestHelper(data *map[tools.OrderedPair]tools.Point, xInfo tools.Info, yInfo tools.Info, proj string, chunk bool) bool {
	chunkR := 3
	chunkC := 2
	pow := 1.7 // don't change

	chunkString := "   "
	outfile := fmt.Sprintf("idw_test_files/idw_step%.0f-%.0f", xInfo.Step, yInfo.Step) // "step{x}-{y}[chunked]pow{pow}.[ext]"
	correctFP := fmt.Sprintf("idw_test_files/idw_correct_step%.0f-%.0f", xInfo.Step, yInfo.Step)

	channel := make(chan string, 2)
	if chunk {
		outfile += "chunked"
		ChunkSolve(data, outfile, xInfo, yInfo, chunkR, chunkC, proj, pow, channel)
	} else {
		chunkString = "NO "
		FullSolve(data, outfile, xInfo, yInfo, proj, pow, false, false, channel)
	}
	outfile += "pow1.7"

	bunyan.Info(<-channel)

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
