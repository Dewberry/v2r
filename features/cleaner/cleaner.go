package cleaner

import (
	"app/tools"
	processing "app/tools/processing"
	"math"

	bunyan "github.com/Dewberry/paul-bunyan"
	// bunyan "github.com/Dewberry/paul-bunyan"
)

func cleanAreaMap(areaMap *[][]square, tolerance map[byte]int, pixelArea float64, adjType int, ICP innerChunkPartition) cleanerStats {
	islands, voids := 0, 0
	islandArea, voidArea := 0, 0

	for r := 0; r < len(*areaMap); r++ {
		for c := 0; c < len((*areaMap)[0]); c++ {
			sq := getSquareRC(areaMap, r, c)
			if sq.IsWater == 255 {
				continue
			}
			if sq.Searched {
				if sq.IsChanged && isInPartiion(ICP, tools.MakePair(r, c)) {
					switch sq.IsWater {
					case byte(0): // from water to land
						voidArea++
						continue

					case byte(1): //land to water
						islandArea++
						continue
					}
				}
				continue
			}

			setSearched(areaMap, tools.MakePair(r, c), true)
			blob, skip := searchBlob(areaMap, tools.MakePair(r, c), adjType, tolerance[getSquareRC(areaMap, r, c).IsWater], getSquareRC(areaMap, r, c).IsWater)
			if skip {
				continue
			}

			if !isBigBlob(&blob) {
				switch blob.IsWater {
				case byte(0): // update island to water
					updateMapFromBlob(areaMap, &blob, 1, true)
					if isInPartiion(ICP, tools.MakePair(r, c)) {
						islands++
						islandArea++
					}
					// islandArea += getNumElements(&blob)

				case byte(1): // update island to water
					updateMapFromBlob(areaMap, &blob, 0, true)
					if isInPartiion(ICP, tools.MakePair(r, c)) {
						voids++
						voidArea++
					}
					// voidArea += getNumElements(&blob)
				}
			}

		}
	}
	return cleanerStats{islands, voids, islandArea, voidArea}
}

func searchBlob(areaMap *[][]square, loc tools.OrderedPair, adjType int, thresholdsize int, wet byte) (blob, bool) {
	blob := blob{make([]tools.OrderedPair, 1, thresholdsize), 0, thresholdsize, wet}
	blob.Elements[0] = loc
	searchStack := []tools.OrderedPair{loc}

	skip := false
	for len(searchStack) > 0 {
		n := len(searchStack) - 1
		searchLoc := searchStack[n]
		searchStack = searchStack[:n]

		adjacents, shouldSkip := getSimilarSurrounding(areaMap, searchLoc, adjType)
		skip = skip || shouldSkip

		for _, adjLoc := range adjacents {
			growBlob(areaMap, &blob, adjLoc)
			searchStack = append(searchStack, adjLoc)
		}
	}
	return blob, skip
}

// returns a list of adjacent locations and whether to skip over blob (reached a finalized location)
// adjacent locations must be unsearched and of similar type (wet/nonwet)
// adjacent locations specified by adjType
func getSimilarSurrounding(areaMap *[][]square, loc tools.OrderedPair, adjType int) ([]tools.OrderedPair, bool) {
	vectors := AdjacentVectors(adjType)
	directions := [2]int{-1, 1}

	skip := false
	var validSurrounding []tools.OrderedPair
	for _, vec := range vectors {
		for _, dir := range directions {
			adjLoc := tools.MakePair(loc.R+dir*vec.R, loc.C+dir*vec.C)
			if isInBounds(areaMap, adjLoc) && sameBlob(areaMap, loc, adjLoc) {
				if getSquarePair(areaMap, adjLoc).Finalized {
					skip = true
				}
				if !beenSearched(areaMap, adjLoc) {
					validSurrounding = append(validSurrounding, adjLoc)
					setSearched(areaMap, adjLoc, true)
				}
			}
		}
	}
	return validSurrounding, skip
}

func CleanFull(filepath string, outfile string, toleranceIsland float64, toleranceVoid float64, adjType int) error {
	areaMap, gdal, err := readFile(filepath)
	bunyan.Infof("[%v, %v]", len(areaMap), len(areaMap[0]))

	if err != nil {
		return err
	}
	areaSize := math.Abs(gdal.XCell * gdal.YCell)
	tolerance := map[byte]int{0: int(toleranceIsland / areaSize), 1: int(toleranceVoid / areaSize)}

	ICP := innerChunkPartition{0, len(areaMap), 0, len(areaMap[0])}
	summary := cleanAreaMap(&areaMap, tolerance, areaSize, adjType, ICP)
	printStats(summary, areaSize)
	return processing.WriteTif(flattenAreaMap(areaMap), gdal, outfile, tools.MakePair(0, 0), tools.MakePair(len(areaMap), len(areaMap[0])), tools.MakePair(len(areaMap), len(areaMap[0])), true)

}
