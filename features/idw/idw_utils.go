package idw

import (
	"app/tools"
	processing "app/tools/processing"
)

func calculateIDW(locs *map[tools.OrderedPair]tools.Point, xInfo tools.Info, yInfo tools.Info, exp float64, r int, c int) tools.Point {
	denom := 0.0

	p0 := tools.RCToPoint(r, c, xInfo, yInfo)

	givenPoint, exists := (*locs)[tools.RCToPair(r, c)]
	if exists {
		return givenPoint
		// *newWeight = givenPoint.Weight
	}

	for _, pointValue := range getInBounds(p0, locs) {
		denomHelper := tools.DistExp(p0, pointValue, exp)

		p0.Weight += pointValue.Weight * denomHelper
		denom += denomHelper

	}
	p0.Weight /= denom
	// *newWeight = p0.Weight
	return p0
}

//Later feature for a bounding method
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

func writeTif(chunk chunkIDW, filename string, gdal processing.GDalInfo, totalSize tools.OrderedPair, i int) {
	grid := chunk.Data
	start := chunk.Pair
	bufferSize := tools.MakePair(len(grid), len(grid[0]))
	processing.WriteFloatTif(flattenGrid(grid), gdal, filename, start, totalSize, bufferSize, i == 0)

}