package tests

import (
	"app/features/idw"
	tools "app/tools"
	processing "app/tools/processing"
	"fmt"
	"os"
	"strings"

	bunyan "github.com/Dewberry/paul-bunyan"
	"github.com/dewberry/gdal"
)

func testIDW() bool {
	bunyan.Info("____________________________")
	bunyan.Info("IDW")
	chunkR := 3
	chunkC := 2
	pow := 1.7 // don't change
	epsg := 2284

	srs := gdal.CreateSpatialReference("")
	err := srs.FromEPSG(epsg)
	if err != nil {
		bunyan.Fatal(err)
	}
	proj, err := srs.ToWKT()
	if err != nil {
		bunyan.Fatal(err)
	}

	filepath := "tests/idw_files/idw_in.txt"
	listPoints, xInfo, yInfo, err := processing.ReadData(filepath)
	if err != nil {
		bunyan.Fatal(err)
	}

	pass := true
	data := tools.MakeCoordSpace(&listPoints, xInfo, yInfo)

	for _, stepxy := range [2][2]float64{{1.0, 1.0}, {2.0, 2.0}} {
		xInfo.Step = stepxy[0]
		yInfo.Step = stepxy[1]

		outfileFull := fmt.Sprintf("tests/idw_files/idw_step%.0f-%.0f", xInfo.Step, yInfo.Step) // "step{x}-{y}[chunked]pow{pow}.[ext]"
		outfileChunk := fmt.Sprintf("%schunked", outfileFull)
		channel := make(chan string, 2)

		go idw.FullSolve(&data, outfileFull, xInfo, yInfo, proj, pow, false, false, channel)
		go idw.ChunkSolve(&data, outfileChunk, xInfo, yInfo, chunkR, chunkC, proj, pow, channel)

		for i := 0; i < 2; i++ {
			<-channel
		}

		completeOutfileTif := fmt.Sprintf("%spow1.7.tiff", outfileFull)
		completeOutfileAsc := fmt.Sprintf("%spow1.7.asc", outfileFull)
		processing.TransferType(completeOutfileTif, completeOutfileAsc, "Int16")

		completeOutfileChunkedTif := fmt.Sprintf("%spow1.7.tiff", outfileChunk)
		completeOutfileChunkedAsc := fmt.Sprintf("%spow1.7.asc", outfileChunk)
		processing.TransferType(completeOutfileChunkedTif, completeOutfileChunkedAsc, "Int16")

		correct := fmt.Sprintf("tests/idw_files/idw_correct_step%.0f-%.0f.asc", xInfo.Step, yInfo.Step)
		bunyan.Infof("     NO CHUNKING: %s         %v", completeOutfileAsc, sameFiles(completeOutfileAsc, correct))
		bunyan.Infof("     CHUNKING: %s     %v", completeOutfileChunkedAsc, sameFiles(completeOutfileChunkedAsc, correct))

		if !sameFiles(completeOutfileAsc, correct) {
			bunyan.Errorf("FILE: %s  | Correct: %s", completeOutfileAsc, correct)
			pass = false
		}
		if !sameFiles(completeOutfileChunkedAsc, correct) {
			bunyan.Errorf("FILE: %s  | Correct: %s", completeOutfileChunkedAsc, correct)
			pass = false
		}

		completeOutfile := completeOutfileTif[:strings.LastIndex(completeOutfileTif, ".tiff")]
		completeOutfileChunked := completeOutfileChunkedTif[:strings.LastIndex(completeOutfileChunkedTif, ".tiff")]
		for _, fp := range [2]string{completeOutfile, completeOutfileChunked} {
			for _, ext := range [3]string{".asc.aux.xml", ".prj", ".tiff"} {
				os.Remove(fp + ext)
			}
		}
	}
	bunyan.Info("____________________________")
	return pass
}
