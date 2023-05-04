package idw

import (
	"github.com/dewberry/v2r/tools"
	"github.com/dewberry/v2r/tools/processing"
)

func calculateIDW(locs *map[tools.OrderedPair]tools.Point, xInfo tools.Info, yInfo tools.Info, exp float64, r int, c int) tools.Point {
	denom := 0.0

	p0 := tools.RCToPoint(r, c, xInfo, yInfo)

	givenPoint, exists := (*locs)[tools.RCToPair(r, c)]
	if exists {
		return givenPoint
	}

	for _, pointValue := range getInBounds(p0, locs) {
		denomHelper := tools.DistExp(p0, pointValue, exp)

		p0.Weight += pointValue.Weight * denomHelper
		denom += denomHelper

	}
	p0.Weight /= denom
	return p0
}

func getInBounds(p tools.Point, data *map[tools.OrderedPair]tools.Point) map[tools.OrderedPair]tools.Point {
	return *data
}

func flattenGrid(grid [][]float64) []float64 {
	unwrappedMatrix := make([]float64, len(grid)*len(grid[0]))
	for r := 0; r < len(grid); r++ {
		for c := 0; c < len(grid[0]); c++ {
			unwrappedMatrix[r*len(grid[0])+c] = grid[r][c]
		}
	}

	return unwrappedMatrix
}

func writeGDAL(chunk chunkIDW, filename string, gdal processing.GDalInfo, totalSize tools.OrderedPair, create bool) error {
	grid := chunk.Data
	start := chunk.Pair
	bufferSize := tools.MakePair(len(grid), len(grid[0]))
	return processing.WriteGDAL(flattenGrid(grid), gdal, filename, start, totalSize, bufferSize, create)
}

func getChannelSize(chunkSize int) int {
	var overhead uint64 = 10000000                  // 10 MB overestimate
	var subprocess uint64 = uint64(chunkSize * 150) // 4-8 bytes per int + actual bytes, 4 stored; overhead per subprocess estimate
	return tools.ChannelSize(subprocess, overhead)
}
