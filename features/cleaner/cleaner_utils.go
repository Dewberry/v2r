package cleaner

import (
	"app/tools"
	processing "app/tools/processing"
	"fmt"
)

type cleanerStats struct {
	Islands    int
	Voids      int
	IslandArea int
	VoidArea   int
}

func printStats(c cleanerStats, pixelArea float64) {
	fmt.Printf("filled in %v islands covering %.2f sq footage\n", c.Islands, float64(c.IslandArea)*pixelArea)
	fmt.Printf("filled in %v voids covering %.2f sq footage\n", c.Voids, float64(c.VoidArea)*pixelArea)
}

func isInPartiion(ICP innerChunkPartition, loc tools.OrderedPair) bool {
	return loc.R >= ICP.RStart && loc.R < ICP.REnd && loc.C >= ICP.CStart && loc.C < ICP.CEnd
}

func (toUpdate *cleanerStats) updateStats(cs cleanerStats) {
	toUpdate.Islands += cs.Islands
	toUpdate.Voids += cs.Voids
	toUpdate.IslandArea += cs.IslandArea
	toUpdate.VoidArea += cs.VoidArea
}

func createAreaMap(flattenedMap []byte, rowsAndCols tools.OrderedPair) [][]square {
	areaMap := make([][]square, rowsAndCols.R)

	for r := 0; r < rowsAndCols.R; r++ {
		areaMap[r] = make([]square, rowsAndCols.C)
		for c := 0; c < rowsAndCols.C; c++ {
			areaMap[r][c].IsWater = flattenedMap[r*rowsAndCols.C+c]
		}
	}
	return areaMap
}

func flattenAreaMap(areaMap [][]square) []byte {
	unwrappedMatrix := make([]byte, len(areaMap)*len(areaMap[0]))
	for r := 0; r < len(areaMap); r++ {
		for c := 0; c < len(areaMap[0]); c++ {
			unwrappedMatrix[r*len(areaMap[0])+c] = areaMap[r][c].IsWater
		}
	}

	return unwrappedMatrix
}

func AdjacentVectors(adjType int) []tools.OrderedPair {
	switch adjType {
	case 4:
		return []tools.OrderedPair{tools.MakePair(0, 1), tools.MakePair(1, 0)}
	case 8:
		return []tools.OrderedPair{tools.MakePair(0, 1), tools.MakePair(1, 0), tools.MakePair(1, 1), tools.MakePair(-1, 1)}
	default:
		return []tools.OrderedPair{tools.MakePair(0, 1), tools.MakePair(1, 0), tools.MakePair(1, 1), tools.MakePair(-1, 1)}
	}
}

func readFileChunk(filepath string, start tools.OrderedPair, size tools.OrderedPair) ([][]square, error) {
	flattenedMap, _, _, err := processing.ReadTif(filepath, start, size, false)
	if err != nil {
		return [][]square{}, err
	}

	return createAreaMap(flattenedMap, size), nil
}

func readFile(filepath string) ([][]square, processing.GDalInfo, error) {
	flattenedMap, gdal, rowsAndCols, err := processing.ReadTif(filepath, tools.MakePair(0, 0), tools.MakePair(0, 0), true)
	if err != nil {
		return [][]square{}, processing.GDalInfo{}, err
	}

	return createAreaMap(flattenedMap, rowsAndCols), gdal, nil

}
