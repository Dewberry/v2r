package tools

import (
	"fmt"
	"math"
	"time"

	"github.com/jmoiron/sqlx"
)

func ReadData(filepath string) (string, error) {
	start := time.Now()

	pointsList, err := ReadIn(filepath)
	if err != nil {
		return "ReadIn() error", err
	}
	data := MakeCoordSpace(pointsList)

	step := .5
	pow := 1.7
	iterations := 1
	channel := make(chan string, iterations)
	// go MainSolve(data, "data/sample", 3.0, true, channel)

	outfile := filepath
	for x := 0; x < iterations; x++ {
		go MainSolve(data, filepath, pow+float64(x)*step, true, outfile, channel)
	}

	for x := 0; x < iterations; x++ {
		received_string := <-channel
		fmt.Println(received_string)
	}

	return fmt.Sprintf("Completed %v iterations in %v", iterations, time.Since(start)), nil
}

func ReadPGData(db *sqlx.DB, query string, xStep float64, yStep float64) (string, error) {

	start := time.Now()
	rows, err := db.Query("SELECT elevation, ST_X(geom), ST_Y(geom) FROM sandbox.location_1;")
	if err != nil {
		return "db query", err
	}

	var db_items []Point
	min_x, min_y := math.Inf(1), math.Inf(1)
	max_x, max_y := math.Inf(-1), math.Inf(-1)

	for rows.Next() {
		var elev, st_x, st_y float64

		err = rows.Scan(&elev, &st_x, &st_y)
		if err != nil {
			return "row scanning error", err
		}

		min_x = math.Min(min_x, st_x)
		max_x = math.Max(max_x, st_x)

		min_y = math.Min(min_y, st_y)
		max_y = math.Max(max_y, st_y)

		db_items = append(db_items, Point{st_x, st_y, elev})
	}
	ConfigureGlobals(min_x, max_x, xStep, min_y, max_y, yStep)
	data := MakeCoordSpace(db_items)

	step := .5
	iterations := 1
	pow := 1.7

	channel := make(chan string, iterations)
	// go MainSolve(data, "data/sample", 3.0, true, channel)
	outfile := fmt.Sprintf("data/step%.0f", CELL)
	for x := 0; x < iterations; x++ {
		go MainSolve(data, "data/sample", pow+float64(x)*step, true, outfile, channel)
	}

	for x := 0; x < iterations; x++ {
		received_string := <-channel
		fmt.Println(received_string)
	}

	return fmt.Sprintf("Completed %v iterations in %v", iterations, time.Since(start)), nil
}
