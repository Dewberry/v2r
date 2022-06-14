package main

import (
	"app/tools"
	"fmt"
	"log"
	"time"
)

func main() {
	start := time.Now()

	//variables to change
	chunkR := 20
	chunkC := 25
	printOut := true
	powStep := .5
	powStart := 1.7 // inclusive
	powStop := 1.7  // inclusive
	stepX := 20.0
	stepY := 20.0
	inputQuery := "SELECT elevation, ST_X(geom), ST_Y(geom) FROM sandbox.location_1;"
	outfileFolder := "data/"
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
		go tools.MainSolve(&data, outfile, pow, printOut, chunkR, chunkC, channel)
	}

	for pow := powStart; pow <= powStop; pow += powStep {
		receivedString := <-channel
		fmt.Println(receivedString)
	}

	fmt.Printf("Completed %v iterations in %v\n", iterations, time.Since(start))

}
