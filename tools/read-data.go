package tools

import (
	"fmt"
	"math"
	"time"

	"github.com/jmoiron/sqlx"
)

func ReadData(filepath string, useChunk bool) (string, error) {
	start := time.Now()

	pointsList, err := ReadIn(filepath)
	if err != nil {
		return "ReadIn() error", err
	}
	data := MakeCoordSpace(&pointsList)

	step := .5
	pow := 1.7
	iterations := 1
	channel := make(chan string, iterations)
	// go MainSolve(data, "data/sample", 3.0, true, channel)

	outfile := filepath
	for x := 0; x < iterations; x++ {
		go MainSolve(&data, filepath, outfile, pow+float64(x)*step, true, useChunk, channel)
	}

	for x := 0; x < iterations; x++ {
		receivedString := <-channel
		fmt.Println(receivedString)
	}

	return fmt.Sprintf("Completed %v iterations in %v - Chunking used: %v", iterations, time.Since(start), useChunk), nil
}

//TODO: change layout | Return:  | change _ to camel case
func ReadPGData(db *sqlx.DB, query string, stepX float64, stepY float64, useChunk bool) (string, error) {

	start := time.Now()
	rows, err := db.Query("SELECT elevation, ST_X(geom), ST_Y(geom) FROM sandbox.location_1;")
	if err != nil {
		return "db query", err
	}

	var listPoints []Point
	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)

	for rows.Next() {
		var elev, x, y float64

		err = rows.Scan(&elev, &x, &y)
		if err != nil {
			return "row scanning error", err
		}

		minX = math.Min(minX, x)
		maxX = math.Max(maxX, x)

		minY = math.Min(minY, y)
		maxY = math.Max(maxY, y)

		listPoints = append(listPoints, Point{x, y, elev})
	}

	//From here
	ConfigureGlobals(minX, maxX, stepX, minY, maxY, stepY)
	data := MakeCoordSpace(&listPoints)

	//TODO move to main as input
	step := .5
	iterations := 1
	pow := 1.7

	channel := make(chan string, iterations)
	// go MainSolve(data, "data/sample", 3.0, true, channel)
	outfile := fmt.Sprintf("data/step%.0f", CELL)
	for x := 0; x < iterations; x++ {
		go MainSolve(&data, "data/sample", outfile, pow+float64(x)*step, false, useChunk, channel)
	}

	for x := 0; x < iterations; x++ {
		receivedString := <-channel
		fmt.Println(receivedString)
	}

	return fmt.Sprintf("Completed %v iterations in %v - Chunking used: %v", iterations, time.Since(start), useChunk), nil
}
