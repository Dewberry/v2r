package tools

import (
	"fmt"
	"math"
	"time"

	"github.com/jmoiron/sqlx"
)

func ReadData(filepath string) error {

	pointsList, err := ReadIn(2, filepath) // 2 dimensions hardcoded
	if err != nil {
		return err
	}

	data := MakeCoordSpace(pointsList, 0, 0, 1, 1)

	channel := make(chan string, 5)
	i := 0
	for pow := 1.0; pow < 3.5; pow += .5 {

		// co := calculateIDW(data, Coord{Point{5, 7, 0}, OrderedPair{5, 7}}, pow)
		// fmt.Println(co)
		// co1 := calculateIDW(data, Coord{Point{10, 7, 0}, OrderedPair{10, 7}}, pow)
		// fmt.Println(co1)
		// co2 := calculateIDW(data, Coord{Point{6, 12, 0}, OrderedPair{6, 12}}, pow)
		// fmt.Println(co2)
		i++
		err = MainSolve(data, filepath, pow, true, channel, 1, 1)
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

	fmt.Println("x:", min_x, max_x, "y:", min_y, max_y)
	fmt.Println("x,", max_x-min_x, "\ty:", max_y-min_y)

	xStep := 10000.0
	yStep := 10000.0

	SetMinMax(min_x, max_x, min_y, max_y, xStep, yStep)
	data := MakeCoordSpace(db_items, min_x, min_y, xStep, yStep)

	// step := .5
	iterations := 1
	// pow := 1.0

	channel := make(chan string, iterations)
	go MainSolve(data, "data/sample", 3.0, true, channel, xStep, yStep)
	// for x := 0; x < iterations; x++ {
	// 	go MainSolve(data, "data/sample", pow+float64(x)*step, true, channel, xStep, yStep)
	// }

	for x := 0; x < iterations; x++ {
		received_string := <-channel
		fmt.Println(received_string)
	}

	return fmt.Sprintf("Completed %v iterations in %v", iterations, time.Since(start)), nil
}
