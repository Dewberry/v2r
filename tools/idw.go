package tools

import (
	"fmt"
	"math"
	"time"
)

var X_MIN_MAX_STEP [3]float64 = [3]float64{0, 0, 1}
var Y_MIN_MAX_STEP [3]float64 = [3]float64{0, 0, 1}
var CELL float64 = 1.0

type ChunkStruct struct {
	Pair  OrderedPair
	Data  [][]float64
	Empty bool
}

func ConfigureGlobals(min_x, max_x, xStep, min_y, max_y, yStep float64) {
	SetStep(math.Sqrt(xStep * yStep)) // QGIS takes square grid
	SetMinMax(min_x, max_x, xStep, min_y, max_y, yStep)
}

func SetMinMax(min_x, max_x, step_x, min_y, max_y, step_y float64) {
	X_MIN_MAX_STEP = [3]float64{min_x, max_x, step_x}
	Y_MIN_MAX_STEP = [3]float64{min_y, max_y, step_y}
}

func SetStep(val float64) {
	CELL = val
}

func ChunkSolve(grid [][]float64, data map[OrderedPair]Point, pow float64) {
	chunkr := 30
	chunkc := 30
	totalChunks := 0
	chunkChannel := make(chan ChunkStruct, 10)
	for r := 0; r < len(grid); r += chunkr {
		for c := 0; c < len(grid[0]); c += chunkc {
			totalChunks++
			go chunkSolveHelper(data, pow, RCToPair(r, c), RCToPair(Min(r+chunkr, len(grid)), Min(c+chunkc, len(grid[0]))), chunkChannel)
		}
	}

	for i := 0; i < totalChunks; i++ {
		gridChunk := <-chunkChannel
		r0, c0 := PairToRC(gridChunk.Pair)

		for r := 0; r < len(gridChunk.Data); r++ {
			for c := 0; c < len(gridChunk.Data[0]); c++ {
				grid[r0+r][c0+c] = gridChunk.Data[r][c]
			}
		}
	}
}

func chunkSolveHelper(locs map[OrderedPair]Point, pow float64, start OrderedPair, end OrderedPair, channel chan ChunkStruct) {
	var grid [][]float64

	empty := true
	rStart, cStart := PairToRC(start)
	rEnd, cEnd := PairToRC(end)
	for r := rStart; r < rEnd; r++ {
		var row []float64
		for c := cStart; c < cEnd; c++ {
			row = append(row, calculateIDW(locs, r, c, pow).Weight)
			empty = false
		}
		grid = append(grid, row)

	}
	channel <- ChunkStruct{start, grid, empty}

}

func MainSolve(data map[OrderedPair]Point, filepath string, outfile string, pow float64, print_out bool, useChunk bool, channel chan string) error {
	start := time.Now()

	xScale := int((1 + X_MIN_MAX_STEP[1] - X_MIN_MAX_STEP[0]) / X_MIN_MAX_STEP[2])
	yScale := int((1 + Y_MIN_MAX_STEP[1] - Y_MIN_MAX_STEP[0]) / Y_MIN_MAX_STEP[2])
	var grid = make([][]float64, yScale)
	for i := range grid {
		grid[i] = make([]float64, xScale)
	}
	fmt.Println("MIN VALUE | MAX VALUE | STEP")
	fmt.Println("X", X_MIN_MAX_STEP)
	fmt.Println("Y: ", Y_MIN_MAX_STEP)

	//Chunk
	if useChunk {
		ChunkSolve(grid, data, pow)
	} else {
		//Unchunk
		for r := 0; r < len(grid); r++ {
			for c := 0; c < len(grid[0]); c++ {
				pt := calculateIDW(data, r, c, pow)
				grid[r][c] = pt.Weight
			}
		}
	}

	start_print := time.Now()

	if print_out {
		// innerErr := PrintExcel(grid, outfile, pow)
		innerErr := PrintAscii(grid, outfile, pow, 1000.0)

		if innerErr != nil {
			return innerErr
		}
	}
	channel <- fmt.Sprintf("pow%v [%vX%v] completed in %v, print took %v", pow, len(grid), len(grid[0]), time.Since(start), time.Since(start_print))
	return nil
}

func calculateIDW(locs map[OrderedPair]Point, r int, c int, exp float64) Point {
	denom := 0.0

	p0 := RCToPoint(r, c)

	p_given, exists := locs[RCToPair(r, c)]
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
