package main

import (
	cleaner "app/features/cleaner"
	idw "app/features/idw"
	"app/tools"
	processing "app/tools/processing"
	"fmt"
	"log"
	"strings"
	"time"
)

func main() {
	doIdw := true
	doClean := false

	if doIdw {
		doIDW()
	}
	if doClean {
		clean()
	}
}

func clean() {
	start := time.Now()

	//variables to change
	adjType := 8 // 4 or 8
	filepath := "data/cleaner/clipped_wet_dry.tif"
	// toleranceIsland := 40000.0 // standard tolerance
	// toleranceVoid := 22500.0   // standard tolerance
	// useChunk := false
	toleranceIsland := 400.0 // test smaller datasets
	toleranceVoid := 225.0   // test smaller datasets
	useChunk := true
	chunkx := 150
	chunky := 100
	chunkChannelSize := 20
	//variables to change

	chunkString := ""
	if useChunk {
		chunkString = "chunked"
	}
	outfile := fmt.Sprintf("%s_isl%.0fvoid%.0f_cleaned%v%v", strings.TrimSuffix(filepath, ".tif"), toleranceIsland, toleranceVoid, adjType, chunkString)

	err := error(nil)
	if useChunk {
		err = cleaner.CleanWithChunking(filepath, outfile, toleranceIsland, toleranceVoid, tools.MakePair(chunky, chunkx), adjType, chunkChannelSize)
	} else {
		err = cleaner.CleanFull(filepath, outfile, toleranceIsland, toleranceVoid, adjType)
	}

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Outfile: %s\nFinished cleaning in %v\n", outfile, time.Since(start))

}

//add chunk printing
func doIDW() {
	start := time.Now()

	//variables to change
	useChunking := true
	chunkR := 200
	chunkC := 250
	powStep := .5
	powStart := 1.7 // inclusive
	powStop := 1.7  // inclusive
	stepX := 100.0
	stepY := 100.0
	epsg := 2284
	inputQuery := "SELECT elevation, ST_X(geom), ST_Y(geom) FROM sandbox.location_1;"
	outfileFolder := "data/idw/"
	//variables to change

	chunkString := ""
	if useChunking {
		chunkString = "chunked"
	}

	// From txt file
	// inputFile := "data/small/nb2.txt"
	// listPoints, xInfo, yInfo, err := processing.ReadData(inputFile)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// From txt file

	//From db
	db := processing.DBInit()

	err := processing.PingWithTimeout(db)
	if err != nil {
		fmt.Println("Connected to database?", err)
	}

	listPoints, xInfo, yInfo, err := processing.ReadPGData(db, inputQuery, stepX, stepY)
	if err != nil {
		log.Fatal(err)
	}
	//From db

	data := tools.MakeCoordSpace(&listPoints, xInfo, yInfo)
	outfile := fmt.Sprintf("%sstep%.0f-%.0f%s", outfileFolder, stepX, stepY, chunkString) // "step{x}-{y}[chunked]pow{pow}.[ext]"
	iterations := 1 + int((powStop-powStart)/powStep)
	channel := make(chan string, iterations)

	for pow := powStart; pow <= powStop; pow += powStep {
		go idw.MainSolve(&data, outfile, xInfo, yInfo, pow, useChunking, chunkR, chunkC, epsg, channel)
	}

	for pow := powStart; pow <= powStop; pow += powStep {
		receivedString := <-channel
		fmt.Println(receivedString)
	}

	fmt.Printf("Completed %v iterations in %v\n", iterations, time.Since(start))
	fmt.Printf("Outfiles: %vpow{x}.tiff\n", outfile)
}
