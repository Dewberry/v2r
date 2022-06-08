package tools

import (
	"fmt"
	"strings"
)

var MIN []int
var MAX []int

func MainSolve(data []Point, filepath string, pow float64, print_out bool) error {

	var grid = make([][]float64, MAX[0]-MIN[0]+1)
	for i := range grid {
		grid[i] = make([]float64, MAX[1]-MIN[1]+1)
	}

	for x := 0; x < len(grid); x++ {
		for y := 0; y < len(grid[0]); y++ {
			grid[x][y] = calculateIDW(data, Point{x + MIN[0], y + MIN[1], 0}, pow).Weight
		}
	}
	filename := fmt.Sprintf("%s-output.xlsx", strings.TrimSuffix(filepath, ".txt"))
	sheetname := fmt.Sprintf("pow%v", pow)

	if print_out {
		innerErr := PrintExcel(grid, filename, sheetname)

		if innerErr != nil {
			return innerErr
		}
	}
	return nil

}

// // currently assumes dimension n = 2
// func makeGrid(grid_min_max [][]float64) [][]float64{

// }

func calculateIDW(data []Point, p0 Point, exp float64) Point {
	denom := 0.0
	for _, p := range getInBounds(p0, data) {
		if p.X == p0.X && p.Y == p0.Y {
			return p
		}
		denom_helper := DistExp(p0, p, exp)

		p0.Weight += p.Weight * denom_helper
		denom += denom_helper

	}
	p0.Weight /= denom
	return p0
}

func getInBounds(p Point, data []Point) []Point {
	return data
}
