package idw

import (
	"app/tools"
	_ "net/http/pprof"
)

type chunkIDW struct {
	Pair tools.OrderedPair
	Data [][]float64
}

func chunkSolve(data *map[tools.OrderedPair]tools.Point, xInfo tools.Info, yInfo tools.Info, exp float64, chunkR int, chunkC int, chunkChannel chan chunkIDW) int {
	numRows, numCols := tools.GetDimensions(xInfo, yInfo)

	totalChunks := 0
	grid := make([][]float64, numRows)
	for r := 0; r < numRows; r++ {
		grid[r] = make([]float64, numCols)
	}

	for r := 0; r < numRows; r += chunkR {
		for c := 0; c < numCols; c += chunkC {
			go makeGridIDW(data, xInfo, yInfo, tools.RCToPair(r, c), tools.RCToPair(tools.Min(r+chunkR, numRows), tools.Min(c+chunkC, numCols)), exp, chunkChannel)

			totalChunks++
		}
	}
	return totalChunks
}

func chunkUpdate(grid *[][]float64, gridChunk *chunkIDW, channel chan bool) {
	r0, c0 := tools.PairToRC(gridChunk.Pair)

	for r := 0; r < len(gridChunk.Data); r++ {
		for c := 0; c < len(gridChunk.Data[0]); c++ {
			(*grid)[r0+r][c0+c] = gridChunk.Data[r][c]
		}
	}

	channel <- true
}

func makeGridIDW(locs *map[tools.OrderedPair]tools.Point, xInfo tools.Info, yInfo tools.Info, start tools.OrderedPair, end tools.OrderedPair, exp float64, channel chan chunkIDW) {
	rStart, cStart := tools.PairToRC(start)
	rEnd, cEnd := tools.PairToRC(end)
	grid := make([][]float64, rEnd-rStart)
	for r := rStart; r < rEnd; r++ {
		grid[r-rStart] = make([]float64, cEnd-cStart)
		for c := cStart; c < cEnd; c++ {
			grid[r-rStart][c-cStart] = calculateIDW(locs, xInfo, yInfo, exp, r, c).Weight
			// calculateIDW(locs, xInfo, yInfo, &grid[r-rStart][c-cStart], exp, r, c)
		}
	}

	channel <- chunkIDW{start, grid}
}
