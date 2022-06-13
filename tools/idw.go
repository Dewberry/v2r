package tools

import (
	"fmt"
	"math"
	"time"
)

var GlobalX [3]float64 = [3]float64{0, 0, 1}
var GlobalY [3]float64 = [3]float64{0, 0, 1}
var CELL float64 = 1.0

type ChunkStruct struct {
	Pair OrderedPair
	Data [][]float64
}

func ConfigureGlobals(minX, maxX, stepX, minY, maxY, stepY float64) {
	SetStep(math.Sqrt(stepX * stepY)) // QGIS takes square grid
	SetMinMax(minX, maxX, stepX, minY, maxY, stepY)
}

func SetMinMax(minX, maxX, stepX, minY, maxY, stepY float64) {
	GlobalX = [3]float64{minX, maxX, stepX}
	GlobalY = [3]float64{minY, maxY, stepY}
}

func SetStep(val float64) {
	CELL = val
}

func ChunkSolve(grid *[][]float64, data *map[OrderedPair]Point, pow float64) {
	chunkr := 1000
	chunkc := 1000
	totalChunks := 0
	chunkChannel := make(chan ChunkStruct, 30)
	for r := 0; r < len((*grid)); r += chunkr {
		for c := 0; c < len((*grid)[0]); c += chunkc {
			totalChunks++
			go chunkSolveHelper(data, pow, RCToPair(r, c), RCToPair(Min(r+chunkr, len(*grid)), Min(c+chunkc, len((*grid)[0]))), chunkChannel)
		}
	}

	for i := 0; i < totalChunks; i++ {
		gridChunk := <-chunkChannel
		go chunkUpdate(grid, &gridChunk)
	}
}

func chunkUpdate(grid *[][]float64, gridChunk *ChunkStruct) {
	r0, c0 := PairToRC(gridChunk.Pair)

	for r := 0; r < len(gridChunk.Data); r++ {
		for c := 0; c < len(gridChunk.Data[0]); c++ {
			(*grid)[r0+r][c0+c] = gridChunk.Data[r][c]
		}
	}
}

func chunkSolveHelper(locs *map[OrderedPair]Point, pow float64, start OrderedPair, end OrderedPair, channel chan ChunkStruct) {
	var grid [][]float64

	rStart, cStart := PairToRC(start)
	rEnd, cEnd := PairToRC(end)
	for r := rStart; r < rEnd; r++ {
		var row []float64
		for c := cStart; c < cEnd; c++ {
			row = append(row, calculateIDW(locs, r, c, pow).Weight)
		}
		grid = append(grid, row)

	}
	channel <- ChunkStruct{start, grid}

}

// (x, y) and r, c change all to r, c
func MainSolve(data *map[OrderedPair]Point, filepath string, outfile string, pow float64, printOut bool, useChunk bool, channel chan string) error {
	start := time.Now()

	xScale := int((1 + GlobalX[1] - GlobalX[0]) / GlobalX[2])
	yScale := int((1 + GlobalY[1] - GlobalY[0]) / GlobalY[2])
	var grid = make([][]float64, yScale)
	for i := range grid {
		grid[i] = make([]float64, xScale)
	}
	fmt.Println("MIN VALUE | MAX VALUE | STEP")
	fmt.Println("X", GlobalX)
	fmt.Println("Y: ", GlobalY)

	//Chunk
	if useChunk {
		ChunkSolve(&grid, data, pow)
	} else {
		//Unchunk
		for r := 0; r < len(grid); r++ {
			for c := 0; c < len(grid[0]); c++ {
				pt := calculateIDW(data, r, c, pow)
				grid[r][c] = pt.Weight
			}
		}
	}

	startPrint := time.Now()

	if printOut {
		// innerErr := PrintExcel(grid, outfile, pow)
		innerErr := PrintAscii(grid, outfile, pow, 1000.0)

		if innerErr != nil {
			return innerErr
		}
	}
	channel <- fmt.Sprintf("pow%v [%vX%v] completed in %v, print took %v", pow, len(grid), len(grid[0]), time.Since(start), time.Since(startPrint))
	return nil
}

func calculateIDW(locs *map[OrderedPair]Point, r int, c int, exp float64) Point {
	denom := 0.0

	p0 := RCToPoint(r, c)

	givenPoint, exists := (*locs)[RCToPair(r, c)]
	if exists {
		return givenPoint
	}

	for _, pointValue := range getInBounds(p0, locs) {
		denomHelper := DistExp(p0, pointValue, exp)

		p0.Weight += pointValue.Weight * denomHelper
		denom += denomHelper

	}
	p0.Weight /= denom
	return p0
}

func getInBounds(p Point, data *map[OrderedPair]Point) map[OrderedPair]Point {
	return *data
}
