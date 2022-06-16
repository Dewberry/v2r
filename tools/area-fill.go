package tools

import (
	"fmt"
	"math"
)

func searchBlob(areaMap *[][]Square, loc OrderedPair, adjType int, thresholdsize int, wet byte) (Blob, bool) {
	blob := Blob{[]OrderedPair{loc}, 0, thresholdsize, wet}
	searchStack := []OrderedPair{loc}

	skip := false
	for len(searchStack) > 0 {
		n := len(searchStack) - 1
		searchLoc := searchStack[n]
		searchStack = searchStack[:n]

		adjacents := getSimilarSurrounding(areaMap, searchLoc, adjType, &skip)

		for _, adjLoc := range adjacents {
			growBlob(areaMap, &blob, adjLoc)
			searchStack = append(searchStack, adjLoc)
		}
	}
	return blob, skip
}

// returns a list of adjacent locations and whether to skip over blob
// adjacent locations must be unsearched and of similar type (wet/nonwet)
// adjacent locations specified by adjType
func getSimilarSurrounding(areaMap *[][]Square, loc OrderedPair, adjType int, skip *bool) []OrderedPair {
	vectors := getVectors(adjType)
	directions := [2]int{-1, 1}

	var validSurrounding []OrderedPair
	for _, vec := range vectors {
		for _, dir := range directions {
			adjLoc := OrderedPair{loc.R + dir*vec.R, loc.C + dir*vec.C}
			if inBounds(areaMap, adjLoc) && sameBlob(areaMap, loc, adjLoc) {
				if getSquarePair(areaMap, adjLoc).Finalized {
					*skip = true
				}
				if !searchedLoc(areaMap, adjLoc) {
					validSurrounding = append(validSurrounding, adjLoc)
					setSearched(areaMap, adjLoc, true)
				}
			}
		}
	}
	return validSurrounding
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

func fillMatrix(areaMap *[][]Square, tolerance map[byte]int, pixelArea float64, adjType int) {
	islands, voids := 0, 0
	islandArea, voidArea := 0, 0

	largest := 0
	fmt.Println(tolerance)
	for r := 0; r < len(*areaMap); r++ {
		for c := 0; c < len((*areaMap)[0]); c++ {
			sq := getSquareRC(areaMap, r, c)
			if sq.Searched || sq.IsWater == 255 {
				continue
			}
			setSearched(areaMap, OrderedPair{r, c}, true)
			blob, skip := searchBlob(areaMap, OrderedPair{r, c}, adjType, tolerance[getSquareRC(areaMap, r, c).IsWater], getSquareRC(areaMap, r, c).IsWater)
			if skip {
				continue
			}

			largest = Max(largest, GetNumElements(&blob))
			if !BigBlob(&blob) {
				switch blob.IsWater {
				case byte(0): // update island to water
					updateMapFromBlob(areaMap, &blob, 1, true)
					islands++
					islandArea += GetNumElements(&blob)

				case byte(1): // update island to water
					updateMapFromBlob(areaMap, &blob, 0, true)
					voids++
					voidArea += GetNumElements(&blob)
				}
			} else {
				fmt.Println("big blob size ", blob.NumFixed)
			}

		}
	}
	fmt.Println(largest)
	fmt.Printf("filled in %v islands covering %.2f sq footage\n", islands, float64(islandArea)*pixelArea)
	fmt.Printf("filled in %v voids covering %.2f sq footage\n", voids, float64(voidArea)*pixelArea)
}

func FullFillMap(filepath string, outfile string, toleranceIsland float64, toleranceVoid float64, adjType int) error {
	areaMap, gdal, err := getMatrixFull(filepath)
	if err != nil {
		return err
	}
	areaSize := math.Abs(gdal.XCell * gdal.YCell)

	fmt.Printf("island: %v\tvoid: %v\t[RowsXCols] [%vX%v]\tarea size: %v\n", toleranceIsland, toleranceVoid, len(areaMap), len(areaMap[0]), areaSize)
	tolerance := map[byte]int{0: int(toleranceIsland / areaSize), 1: int(toleranceVoid / areaSize)}
	fillMatrix(&areaMap, tolerance, areaSize, adjType)
	return WriteTifSquare(areaMap, gdal, outfile)

}

func getMatrixFull(filepath string) ([][]Square, GDalInfo, error) {
	flattenedMap, gdal, rowsAndCols, err := ReadTif(filepath)
	if err != nil {
		return [][]Square{}, GDalInfo{}, err
	}

	return makeAreaMap(flattenedMap, rowsAndCols), gdal, nil

}

func ChunkFillMap(filepath string, outfile string, toleranceIsland float64, toleranceVoid float64, adjType int, chunkSize OrderedPair) error {
	return nil

}

func getMatrixChunk(filepath string, chunkSize OrderedPair) {
	return
}
