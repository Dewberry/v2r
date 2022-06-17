package main

import (
	"app/tools"
	"fmt"
	"log"
	"strings"
	"time"
)

func main() {
	// doIDW()
	fill()
}

func fill() {
	start := time.Now()

	//variables to change
	adjType := 8 // 4 or 8
	filepath := "data/cleaner/clipped_wet_dry.tif"
	// toleranceIsland := 40000.0
	// toleranceVoid := 22500.0
	// useChunk := false
	toleranceIsland := 400.0
	toleranceVoid := 225.0
	useChunk := true
	chunkx := 150
	chunky := 100
	chunkChannelSize := 20

	//variables to change
	chunkString := ""
	if useChunk {
		chunkString = "chunked"
	}
	outfile := fmt.Sprintf("%s_filled%v%v", strings.TrimSuffix(filepath, ".tif"), adjType, chunkString)

	err := error(nil)
	if useChunk {
		err = tools.ChunkFillMap(filepath, outfile, toleranceIsland, toleranceVoid, tools.MakePair(chunky, chunkx), adjType, chunkChannelSize)
	} else {
		err = tools.FullFillMap(filepath, outfile, toleranceIsland, toleranceVoid, adjType)
	}

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Finished fill in %v\n", time.Since(start))

}

//add chunk printing
func doIDW() {
	start := time.Now()

	//variables to change
	chunkR := 200
	chunkC := 250
	printOut := true
	powStep := .5
	powStart := 1.7 // inclusive
	powStop := 1.7  // inclusive
	stepX := 10.0
	stepY := 10.0
	espg := 2284
	inputQuery := "SELECT elevation, ST_X(geom), ST_Y(geom) FROM sandbox.location_1;"
	outfileFolder := "data/idw/"
	//variables to change

	// From txt file
	// inputFile := "data/small/nb2.txt"
	// listPoints, err := tools.ReadData(inputFile, useChunk)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// From txt file

	//From db
	db := tools.DBInit()

	err := tools.PingWithTimeout(db)
	if err != nil {
		fmt.Println("Connected to database?", err)
	}

	listPoints, err := tools.ReadPGData(db, inputQuery, stepX, stepY)
	if err != nil {
		log.Fatal(err)
	}
	//From db

	data := tools.MakeCoordSpace(&listPoints)

	outfile := fmt.Sprintf("%sstep%.0f-%.0f", outfileFolder, stepX, stepY) // "step{x}-{y}pow{pow}.asc"
	iterations := 1 + int((powStop-powStart)/powStep)
	channel := make(chan string, iterations)
	for pow := powStart; pow <= powStop; pow += powStep {
		go tools.MainSolve(&data, outfile, pow, printOut, chunkR, chunkC, espg, channel)
	}

	for pow := powStart; pow <= powStop; pow += powStep {
		receivedString := <-channel
		fmt.Println(receivedString)
	}

	fmt.Printf("Completed %v iterations in %v\n", iterations, time.Since(start))
}
