package cleaner

import (
	"app/tools"
	processing "app/tools/processing"
	"fmt"
	"math"
)

func cleanAreaMap(areaMap *[][]square, tolerance map[byte]int, pixelArea float64, adjType int, verbose bool) {
	islands, voids := 0, 0
	islandArea, voidArea := 0, 0

	for r := 0; r < len(*areaMap); r++ {
		for c := 0; c < len((*areaMap)[0]); c++ {
			sq := getSquareRC(areaMap, r, c)
			if sq.Searched || sq.IsWater == 255 {
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
					islands++
					islandArea += getNumElements(&blob)

				case byte(1): // update island to water
					updateMapFromBlob(areaMap, &blob, 0, true)
					voids++
					voidArea += getNumElements(&blob)
				}
			}

		}
	}
	if verbose { //logging paul-bunyan
		fmt.Printf("filled in %v islands covering %.2f sq footage\n", islands, float64(islandArea)*pixelArea)
		fmt.Printf("filled in %v voids covering %.2f sq footage\n", voids, float64(voidArea)*pixelArea)
	}
}

func searchBlob(areaMap *[][]square, loc tools.OrderedPair, adjType int, thresholdsize int, wet byte) (blob, bool) {
	blob := blob{[]tools.OrderedPair{loc}, 0, thresholdsize, wet}
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
	if err != nil {
		return err
	}
	areaSize := math.Abs(gdal.XCell * gdal.YCell)
	tolerance := map[byte]int{0: int(toleranceIsland / areaSize), 1: int(toleranceVoid / areaSize)}

	cleanAreaMap(&areaMap, tolerance, areaSize, adjType, true)
	return processing.WriteByteTif(flattenAreaMap(areaMap), gdal, tools.MakePair(0, 0), tools.MakePair(len(areaMap), len(areaMap[0])), tools.MakePair(len(areaMap), len(areaMap[0])), outfile, true)
}
