package tools

import (
	"fmt"
	"time"
)

var X_MIN_MAX_STEP [3]float64 = [3]float64{0, 0, 1}
var Y_MIN_MAX_STEP [3]float64 = [3]float64{0, 0, 1}

type ChunkStruct struct {
	pair OrderedPair
	data [][]float64
}

func MainSolve(data map[OrderedPair]Point, filepath string, pow float64, print_out bool, channel chan string) error {
	start := time.Now()

	xScale := int((1 + X_MIN_MAX_STEP[1] - X_MIN_MAX_STEP[0]) / X_MIN_MAX_STEP[2])
	yScale := int((1 + Y_MIN_MAX_STEP[1] - Y_MIN_MAX_STEP[0]) / Y_MIN_MAX_STEP[2])
	var grid = make([][]float64, xScale)
	for i := range grid {
		grid[i] = make([]float64, yScale)
	}
	fmt.Println("MIN VALUE | MAX VALUE | STEP")
	fmt.Println("X", X_MIN_MAX_STEP)
	fmt.Println("Y: ", Y_MIN_MAX_STEP)

	// chunkx := 30
	// chunky := 30
	// totalChunks := 0
	// chunkChannel := make(chan ChunkStruct, 20)
	// for x := 0; x < len(grid); x+=chunkx {
	// 	for y := 0; y < len(grid[0]); y+=chunky {
	// 		totalChunks++
	// 		chunkSolve(data, pow, chunkChannel, OrderedPair{x, y}, OrderedPair{math.MinInt(x+chunkx, len(grid)), math.MinInt(y+chunky, len(grid[0]))})
	// 	}
	// }

	// for i := 0; i < totalChunks; i++ {
	// 	inChunk := <- chunkChannel

	// 	for r:= 0; r <
	// }
	for x := 0; x < len(grid); x++ {
		for y := 0; y < len(grid[0]); y++ {
			px := X_MIN_MAX_STEP[0] + float64(x)*X_MIN_MAX_STEP[2]
			py := Y_MIN_MAX_STEP[0] + float64(y)*Y_MIN_MAX_STEP[2]
			point := Point{px, py, 0}
			// pair := PointToPair(point)

			co := calculateIDW(data, point, pow)
			grid[x][y] = co.Weight
		}
	}
	fmt.Printf("idw finished: %v\n", time.Since(start))

	start_print := time.Now()

	if print_out {
		// innerErr := PrintExcel(grid, filepath, pow)
		innerErr := PrintAscii(grid, filepath, pow, 100.0)

		if innerErr != nil {
			return innerErr
		}
	}
	channel <- fmt.Sprintf("pow%v [%vX%v] completed in %v, print took %v", pow, len(grid), len(grid[0]), time.Since(start), time.Since(start_print))
	return nil

}

func chunkSolve(locs map[OrderedPair]Point, pow float64, channel chan ChunkStruct, start OrderedPair, end OrderedPair) {
	var grid [][]float64
	for r := start.X; r < end.X; r++ {
		var row []float64
		for c := start.Y; c < end.Y; c++ {
			px := X_MIN_MAX_STEP[0] + float64(r)*X_MIN_MAX_STEP[2]
			py := Y_MIN_MAX_STEP[0] + float64(c)*Y_MIN_MAX_STEP[2]
			point := Point{px, py, 0}
			row = append(row, calculateIDW(locs, point, pow).Weight)
		}
		grid = append(grid, row)

	}
	channel <- ChunkStruct{start, grid}

}

func calculateIDW(locs map[OrderedPair]Point, p0 Point, exp float64) Point {
	denom := 0.0

	p_given, exists := locs[PointToPair(p0)]
	if exists {
		return p_given
	}

	for _, val_point := range getInBounds(p0, locs) { // refactor into a map rather than a slice
		denom_helper := DistExp(p0, val_point, exp)

		p0.Weight += val_point.Weight * denom_helper
		denom += denom_helper

	}
	p0.Weight /= denom
	return p0
}

func getInBounds(p Point, data map[OrderedPair]Point) map[OrderedPair]Point {
	return data
}
