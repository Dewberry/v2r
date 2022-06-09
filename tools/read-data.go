package tools

import (
	"fmt"
	"math"
	"time"

	"github.com/jmoiron/sqlx"
)

func ReadData(filepath string) error {

	pointsList, err := ReadIn(filepath)
	if err != nil {
		return err
	}
	data := MakeCoordSpace(pointsList)

	channel := make(chan string, 6)
	i := 0
	for pow := 1.0; pow < 3.5; pow += .5 {

		i++
		err = MainSolve(data, filepath, pow, true, channel)
		if err != nil {
			return err
		}
	}

	for x := 0; x < i; x++ {
		received_string := <-channel
		fmt.Println(received_string)
	}

	return nil
}

func ReadPGData(db *sqlx.DB, query string) (string, error) {

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

	xStep := 1000.0
	yStep := 1000.0

	SetMinMax(min_x, max_x, xStep, min_y, max_y, yStep)
	data := MakeCoordSpace(db_items)

	step := .5
	iterations := 1
	pow := 1.7

	channel := make(chan string, iterations)
	// go MainSolve(data, "data/sample", 3.0, true, channel)
	for x := 0; x < iterations; x++ {
		go MainSolve(data, "data/sample", pow+float64(x)*step, true, channel)
	}

	for x := 0; x < iterations; x++ {
		received_string := <-channel
		fmt.Println(received_string)
	}

	return fmt.Sprintf("Completed %v iterations in %v", iterations, time.Since(start)), nil
}
