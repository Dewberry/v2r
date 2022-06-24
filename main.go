package main

import (
	cleaner "app/features/cleaner"
	idw "app/features/idw"
	"app/tests"
	"app/tools"
	processing "app/tools/processing"
	"flag"
	"fmt"
	"strings"
	"time"

	bunyan "github.com/Dewberry/paul-bunyan"
	"github.com/dewberry/gdal"
)

func main() {
	var path string
	var runIDW bool
	var runClean bool
	var runTests bool
	var runDebug bool
	var runError bool
	flag.StringVar(&path, "f", "data/cleaner/clipped_wet_dry.tif", "pathtotif")
	flag.BoolVar(&runIDW, "i", false, "run the idw?")
	flag.BoolVar(&runClean, "c", false, "run the cleaner?")
	flag.BoolVar(&runTests, "t", false, "run the tests?")
	flag.BoolVar(&runDebug, "d", false, "run the tests?")
	flag.BoolVar(&runError, "e", false, "run the tests?")
	flag.Parse()

	if runError {
		logger := bunyan.New()
		logger.SetLevel(bunyan.ERROR)
	} else if runDebug {
		tools.SetLogging(bunyan.DEBUG)
	} else {
		logger := bunyan.New()
		logger.SetLevel(bunyan.INFO)
	}
	// bunyan.Info("path: ", path)

	if runIDW {
		bunyan.Info("IDW started")
		doIDW()
	}
	if runClean {
		bunyan.Info("Cleaner started")
		clean(path)
	}
	if runTests {
		bunyan.Info("Test suite started")
		tests.TestSuite()
	}
	fmt.Println("program ended")
}

func clean(filepath string) {
	start := time.Now()

	//variables to change
	adjType := 8 // 4 or 8
	// filepath := "data/cleaner/WD_2100_MHHW.tif" // passed through
	// toleranceIsland := 40000.0 // standard tolerance
	// toleranceVoid := 22500.0   // standard tolerance
	// useChunk := false
	toleranceIsland := 40000.0 // test smaller datasets
	toleranceVoid := 22500.0   // test smaller datasets
	useChunk := true
	chunkx := 256 * 10
	chunky := 256 * 10
	// chunkx := 100
	// chunky := 100
	//variables to change

	chunkString := ""
	if useChunk {
		chunkString = "chunked"
	}
	outfile := fmt.Sprintf("%s_isl%.0fvoid%.0f_cleaned%v%v", strings.TrimSuffix(filepath, ".tif"), toleranceIsland, toleranceVoid, adjType, chunkString)

	err := error(nil)
	if useChunk {
		err = cleaner.CleanWithChunking(filepath, outfile, toleranceIsland, toleranceVoid, tools.MakePair(chunky, chunkx), adjType)
	} else {
		err = cleaner.CleanFull(filepath, outfile, toleranceIsland, toleranceVoid, adjType)
	}

	if err != nil {
		bunyan.Fatal(err)
	}

	bunyan.Infof("Outfile: %s\nFinished cleaning in %v\n", outfile, time.Since(start))

}

//add chunk printing
func doIDW() {
	start := time.Now()

	//variables to change
	useChunking := true
	chunkR := 200 * 10
	chunkC := 200 * 10
	powStep := .5
	powStart := 1.7 // inclusive
	powStop := 1.7  // inclusive
	stepX := 10.0
	stepY := 10.0
	epsg := 2284
	outfileFolder := "data/idw/"
	//variables to change

	srs := gdal.CreateSpatialReference("")
	err := srs.FromEPSG(epsg)
	if err != nil {
		bunyan.Fatal(err)
	}
	proj, err := srs.ToWKT()
	if err != nil {
		bunyan.Fatal(err)
	}

	chunkString := ""
	if useChunking {
		chunkString = "chunked"
	}

	// From txt file
	// inputFile := "data/small/idw_in.txt"
	// listPoints, xInfo, yInfo, err := processing.ReadData(inputFile)
	// if err != nil {
	// 	bunyan.Fatal(err)
	// }
	// From txt file

	//From db
	db := processing.DBInit()
	inputQuery := "SELECT elevation, ST_X(geom), ST_Y(geom) FROM sandbox.location_1;"
	err = processing.PingWithTimeout(db)
	if err != nil {
		fmt.Println("Connected to database?", err)
	}

	listPoints, xInfo, yInfo, err := processing.ReadPGData(db, inputQuery, stepX, stepY)
	if err != nil {
		bunyan.Fatal(err)
	}
	//From db

	data := tools.MakeCoordSpace(&listPoints, xInfo, yInfo)
	outfile := fmt.Sprintf("%sstep%.0f-%.0f%s", outfileFolder, stepX, stepY, chunkString) // "step{x}-{y}[chunked]pow{pow}.[ext]"
	iterations := 1 + int((powStop-powStart)/powStep)
	channel := make(chan string, iterations)

	for pow := powStart; pow <= powStop; pow += powStep {
		if !useChunking {
			go idw.FullSolve(&data, outfile, xInfo, yInfo, proj, pow, channel)
		} else {
			go idw.ChunkSolve(&data, outfile, xInfo, yInfo, chunkR, chunkC, proj, pow, channel)
		}
	}

	for pow := powStart; pow <= powStop; pow += powStep {
		receivedString := <-channel
		bunyan.Infof(receivedString)
	}

	bunyan.Infof("Completed %v iterations in %v\n", iterations, time.Since(start))
	bunyan.Infof("Outfiles: %vpow{x}\n", outfile)
}
