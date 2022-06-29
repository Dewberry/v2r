package tests

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/dewberry/v2r/features/idw"
	"github.com/dewberry/v2r/tools"
	"github.com/dewberry/v2r/tools/processing"

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

func isListCorrect(listPoints []tools.Point) bool {
	correct := []tools.Point{
		tools.MakePoint(1.203154304903684e+07, 3.953694544257536e+06, 0),
		tools.MakePoint(1.2039094668343753e+07, 3.9424537521008854e+06, 0),
		tools.MakePoint(1.2052295156865465e+07, 3.93148004717589e+06, 0),
		tools.MakePoint(1.2076264960092723e+07, 3.917394837044359e+06, 0),
		tools.MakePoint(1.209133966962186e+07, 3.899012514405147e+06, 0),
		tools.MakePoint(1.2020401943588393e+07, 3.9463317049700455e+06, 2.5),
		tools.MakePoint(1.2025601603362886e+07, 3.9380775540176313e+06, 2.9),
		tools.MakePoint(1.203495724450095e+07, 3.92078279952575e+06, 7),
		tools.MakePoint(1.2055886381444821e+07, 3.898699015764322e+06, 6.3),
		tools.MakePoint(1.207451364898982e+07, 3.885548111378168e+06, 2),
		tools.MakePoint(1.2049260510355683e+07, 3.952539493517972e+06, 3.1),
		tools.MakePoint(1.205195120503224e+07, 3.947886314968573e+06, 2.8),
		tools.MakePoint(1.2096825067172093e+07, 3.9268244968121317e+06, 3.9),
		tools.MakePoint(1.2080323349286443e+07, 3.9412021225619004e+06, 4.6)}

	if len(correct) != len(listPoints) {
		return false
	}

	for i, p := range correct {
		if p != listPoints[i] {
			return false
		}
	}

	return true
}

func correctSRS(srs gdal.SpatialReference) bool {
	filename := "tests/idw_files/gpkg_proj.txt"
	f, err := os.Open(filename)
	if err != nil {
		bunyan.Fatal(err)
	}
	sc := bufio.NewScanner(f)
	defer f.Close()

	sc.Scan()
	correctProj := sc.Text()
	createdProj, err := srs.ToWKT()
	if err != nil {
		bunyan.Fatal("error in correctSRS() test")

	}
	return correctProj == createdProj
}

func correctXInfo(xInfo tools.Info) bool {
	return xInfo.Min == 1.2020401943588393e+07 && xInfo.Max == 1.2096825067172093e+07 && xInfo.Step == 1.0
}

func correctYInfo(yInfo tools.Info) bool {
	return yInfo.Min == 3.885548111378168e+06 && yInfo.Max == 3.953694544257536e+06 && yInfo.Step == 2.0
}

func testGPKGRead() bool {
	bunyan.Info("____________________________")
	bunyan.Info("read gpkg")

	filepath := "tests/idw_files/input.gpkg"
	listPoints, srs, xInfo, yInfo := processing.ReadGeoPackage(filepath, "wsels", "elevation", 1.0, 2.0)

	if !isListCorrect(listPoints) {
		bunyan.Error("points generated from geopackage incorrect")
	}
	if !correctSRS(srs) {
		bunyan.Error("srs generated incorrect")
	}
	if !correctXInfo(xInfo) {
		bunyan.Error("incorrect xInfo generated")
	}
	if !correctYInfo(yInfo) {
		bunyan.Error("incorrect yInfo generated")
	}

	bunyan.Debug(listPoints, srs, xInfo, yInfo)
	bunyan.Info("____________________________")

	return true
}
