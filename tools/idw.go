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
	ChunkNum int
	Data     *[][]float64
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

func ChunkSolve(data *map[OrderedPair]Point, pow float64, chunkR int, chunkC int) *[]*[][]float64 {
	numRows, numCols := GetDimensions()

	totalChunks := 0
	chunkChannel := make(chan ChunkStruct, 10000)
	for r := 0; r < numRows; r += chunkR {
		for c := 0; c < numCols; c += chunkC {
			go chunkSolveHelper(data, pow, RCToPair(r, c), RCToPair(Min(r+chunkR, numRows), Min(c+chunkC, numCols)), totalChunks, chunkChannel)
			totalChunks++
		}
	}
	partitionedGrid := make([]*[][]float64, totalChunks)
	// fmt.Println("total chunks:", totalChunks)

	for i := 0; i < totalChunks; i++ {
		gridChunk := <-chunkChannel
		chunkUpdate(&partitionedGrid, &gridChunk)
	}
	return &partitionedGrid
}

func chunkUpdate(partitionedGrid *[]*[][]float64, gridChunk *ChunkStruct) {
	(*partitionedGrid)[gridChunk.ChunkNum] = gridChunk.Data
}

func chunkSolveHelper(locs *map[OrderedPair]Point, pow float64, start OrderedPair, end OrderedPair, chunkNum int, channel chan ChunkStruct) {
	rStart, cStart := PairToRC(start)
	rEnd, cEnd := PairToRC(end)
	grid := make([][]float64, rEnd-rStart)
	for r := rStart; r < rEnd; r++ {
		row := make([]float64, cEnd-cStart)
		for c := cStart; c < cEnd; c++ {
			row[c-cStart] = calculateIDW(locs, r, c, pow).Weight
		}
		grid[r-rStart] = row
	}

	channel <- ChunkStruct{chunkNum, &grid}
}

// (x, y) and r, c change all to r, c
func MainSolve(data *map[OrderedPair]Point, outfile string, pow float64, printOut bool, chunkR int, chunkC int, channel chan string) error {
	start := time.Now()

	numRows, numCols := GetDimensions()
	// var grid = make([][]float64, yScale)
	// for i := range grid {
	// 	grid[i] = make([]float64, xScale)
	// }

	grid := ChunkSolve(data, pow, chunkR, chunkC)

	//Chunk
	// if useChunk {
	// 	ChunkSolve(&grid, data, pow)
	// }
	// else {
	// 	//Unchunk
	// 	for r := 0; r < len(grid); r++ {
	// 		for c := 0; c < len(grid[0]); c++ {
	// 			pt := calculateIDW(data, r, c, pow)
	// 			grid[r][c] = pt.Weight
	// 		}
	// 	}
	// }

	startPrint := time.Now()

	if printOut {
		// innerErr := PrintExcel(grid, fmt.Sprintf("%spow%.1f", outfile, pow), pow)
		// innerErr := PrintAscii(grid, fmt.Sprintf("%spow%.1f", outfile, pow), pow, chunkR, chunkC)

		// if innerErr != nil {
		// 	return innerErr
		// }
	}
	if 1 == 2 {
		fmt.Println(grid)
	}
	channel <- fmt.Sprintf("pow%v [%vX%v] completed in %v, print took %v", pow, numRows, numCols, time.Since(start), time.Since(startPrint))
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
