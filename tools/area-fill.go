package tools

import (
	"fmt"
	"math"
)

func searchBlob(areaMap *[][]Square, loc OrderedPair, adjType int) (Blob, bool) {
	blob := Blob{[]OrderedPair{loc}, getSquarePair(areaMap, loc).IsWater}
	searchStack := []OrderedPair{loc}

	for len(searchStack) > 0 {
		n := len(searchStack) - 1
		searchLoc := searchStack[n]
		searchStack = searchStack[:n]

		adjacents, skip := getSimilarSurrounding(areaMap, searchLoc, adjType)
		if skip {
			return Blob{}, true
		}
		for _, adjLoc := range adjacents {
			growBlob(&blob, adjLoc)
			searchStack = append(searchStack, adjLoc)
		}
	}
	return blob, false
}

// returns a list of adjacent locations and whether to skip over blob
// adjacent locations must be unsearched and of similar type (wet/nonwet)
// adjacent locations specified by adjType
func getSimilarSurrounding(areaMap *[][]Square, loc OrderedPair, adjType int) ([]OrderedPair, bool) {
	vectors := getVectors(adjType)
	directions := [2]int{-1, 1}

	var validSurrounding []OrderedPair
	for _, vec := range vectors {
		for _, dir := range directions {
			adjLoc := OrderedPair{loc.R + dir*vec.R, loc.C + dir*vec.C}
			if inBounds(areaMap, adjLoc) && sameBlob(areaMap, loc, adjLoc) {
				if getSquarePair(areaMap, adjLoc).Finalized {
					return []OrderedPair{}, true
				}
				if !searchedLoc(areaMap, adjLoc) {
					validSurrounding = append(validSurrounding, adjLoc)
					setSearched(areaMap, adjLoc, true)
				}
			}
		}
	}
	return validSurrounding, false
}

func getVectors(adjType int) []OrderedPair {
	switch adjType {
	case 4:
		return []OrderedPair{{0, 1}, {1, 0}}
	case 8:
		return []OrderedPair{{0, 1}, {1, 0}, {1, 1}, {-1, 1}}
	default:
		return []OrderedPair{{0, 1}, {1, 0}, {1, 1}, {-1, 1}}
	}
}

func makeAreaMap(flattenedMap []byte, rowsAndCols OrderedPair) [][]Square {
	areaMap := make([][]Square, rowsAndCols.R)

	for r := 0; r < rowsAndCols.R; r++ {
		areaMap[r] = make([]Square, rowsAndCols.C)
		for c := 0; c < rowsAndCols.C; c++ {
			areaMap[r][c].IsWater = flattenedMap[r*rowsAndCols.C+c]
		}
	}
	return areaMap
}

func fillMatrix(areaMap *[][]Square, toleranceIsland float64, toleranceVoid float64, pixelArea float64, adjType int) {
	islands, voids := 0, 0
	islandArea, voidArea := 0, 0
	for r := 0; r < len(*areaMap); r++ {
		for c := 0; c < len((*areaMap)[0]); c++ {
			sq := getSquareRC(areaMap, r, c)
			if sq.Searched || sq.IsWater == 255 {
				continue
			}
			setSearched(areaMap, OrderedPair{r, c}, true)
			blob, skip := searchBlob(areaMap, OrderedPair{r, c}, adjType)
			if skip {
				continue
			}

			if blob.IsWater == 1 { // water - void
				if pixelArea*float64(len(blob.Elements)) <= toleranceVoid {
					updateMapFromBlob(areaMap, &blob, 0)
					voids++
					voidArea += len(blob.Elements)
				} else {
					updateMapFromBlob(areaMap, &blob, 1) // finalize location
				}
			} else if blob.IsWater == 0 { // land - island
				if pixelArea*float64(len(blob.Elements)) <= toleranceIsland {
					updateMapFromBlob(areaMap, &blob, 1)
					islands++
					islandArea += len(blob.Elements)
				} else {
					updateMapFromBlob(areaMap, &blob, 0) // finalize location
				}
			}

		}
	}
	fmt.Printf("filled in %v islands covering %.2f sq footage\n", islands, float64(islandArea)*pixelArea)
	fmt.Printf("filled in %v voids covering %.2f sq footage\n", voids, float64(voidArea)*pixelArea)
}

func AreaFill(filepath string, outfile string, toleranceIsland float64, toleranceVoid float64, adjType int) error {
	flattenedMap, gdal, rowsAndCols, err := ReadTif(filepath)
	if err != nil {
		return err
	}

	fmt.Printf("island: %v\tvoid: %v\t[%vX%v]\tarea size: %v\n", toleranceIsland, toleranceVoid, rowsAndCols.R, rowsAndCols.C, math.Abs(gdal.XCell*gdal.YCell))
	areaMap := makeAreaMap(flattenedMap, rowsAndCols)
	fillMatrix(&areaMap, toleranceIsland, toleranceVoid, math.Abs(gdal.XCell*gdal.YCell), adjType)

	return WriteTifSquare(areaMap, gdal, outfile)

}
