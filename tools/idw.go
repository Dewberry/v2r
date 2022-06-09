package tools

import (
	"fmt"
	"strings"
	"time"
)

var MIN = make([]int, 2)
var MAX = make([]int, 2)

func MainSolve(data map[OrderedPair]Point, filepath string, pow float64, print_out bool, channel chan string, xStep float64, yStep float64) error {

	start := time.Now()
	var grid = make([][]float64, MAX[0]-MIN[0]+1)
	for i := range grid {
		grid[i] = make([]float64, MAX[1]-MIN[1]+1)
	}
	fmt.Println("MAX", MAX, "MIN", MIN)

	for x := 0; x < len(grid); x++ {
		for y := 0; y < len(grid[0]); y++ {
			fmt.Println("coord", Coord{Point{float64(MIN[0]) + float64(x)*xStep, float64(MIN[1]) + float64(y)*yStep, 0}, OrderedPair{x + MIN[0], y + MIN[1]}})
			co := calculateIDW(data, Coord{Point{float64(MIN[0]) + float64(x)*xStep, float64(MIN[1]) + float64(y)*yStep, 0}, OrderedPair{x + MIN[0], y + MIN[1]}}, pow)
			// fmt.Println(co)
			grid[x][y] = co.P.Weight
		}
	}
	filename := fmt.Sprintf("%s-output.xlsx", strings.TrimSuffix(filepath, ".txt"))
	sheetname := fmt.Sprintf("pow%v", pow)

	start_print := time.Now()

	if print_out {
		innerErr := PrintExcel(grid, filename, sheetname)

		if innerErr != nil {
			return innerErr
		}
	}
	channel <- fmt.Sprintf("pow%v [%vX%v] completed in %v, print took %v", pow, len(grid), len(grid[0]), time.Since(start), time.Since(start_print))
	return nil

}

// // currently assumes dimension n = 2
// func makeGrid(grid_min_max [][]float64) [][]float64{

// }

func calculateIDW(locs map[OrderedPair]Point, p0 Coord, exp float64) Coord {
	denom := 0.0

	point, exists := locs[p0.Pair]
	if exists {
		return Coord{point, p0.Pair}
	}

	//(x, y) : (actualx, actualy, weight)
	// fmt.Println(p0)
	for _, val_point := range getInBounds(p0, locs) { // refactor into a map rather than a slice
		denom_helper := DistExp(p0.P, val_point, exp)
		// fmt.Println(denom_helper, val_point)

		p0.P.Weight += val_point.Weight * denom_helper
		denom += denom_helper

	}
	p0.P.Weight /= denom
	return p0
}

func getInBounds(p Coord, data map[OrderedPair]Point) map[OrderedPair]Point {
	return data
}
