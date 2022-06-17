package tools

import (
	"fmt"
	"log"
	"math"
)

type chunkFillStruct struct {
	AreaMap [][]Square
	Offset  OrderedPair
}

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

func sliceAreaMap(areaMap *[][]Square, valid [4]int) {

	// fmt.Printf("valid: %v\tprev dim [%vX%v]", valid, len(*areaMap), len((*areaMap)[0]))
	*areaMap = (*areaMap)[valid[0]:valid[1]]
	for r := 0; r < len(*areaMap); r++ {
		(*areaMap)[r] = (*areaMap)[r][valid[2]:valid[3]]
	}
	// fmt.Printf("\tnew dim [%vX%v]\n", len(*areaMap), len((*areaMap)[0]))
}

func flattenAreaMap(areaMap [][]Square) []byte {
	unwrappedMatrix := make([]byte, len(areaMap)*len(areaMap[0]))
	for r := 0; r < len(areaMap); r++ {
		for c := 0; c < len(areaMap[0]); c++ {
			unwrappedMatrix[r*len(areaMap[0])+c] = areaMap[r][c].IsWater
		}
	}

	return unwrappedMatrix
}

func FullFillMatrix(areaMap *[][]Square, tolerance map[byte]int, pixelArea float64, adjType int, verbose bool) {
	islands, voids := 0, 0
	islandArea, voidArea := 0, 0

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
			}

		}
	}
	if verbose {
		fmt.Printf("filled in %v islands covering %.2f sq footage\n", islands, float64(islandArea)*pixelArea)
		fmt.Printf("filled in %v voids covering %.2f sq footage\n", voids, float64(voidArea)*pixelArea)
	}
}

func FullFillMap(filepath string, outfile string, toleranceIsland float64, toleranceVoid float64, adjType int) error {
	areaMap, gdal, err := getMatrixFull(filepath)
	if err != nil {
		return err
	}
	areaSize := math.Abs(gdal.XCell * gdal.YCell)

	// fmt.Printf("island: %v\tvoid: %v\t[RowsXCols] [%vX%v]\tarea size: %v\n", toleranceIsland, toleranceVoid, len(areaMap), len(areaMap[0]), areaSize)
	tolerance := map[byte]int{0: int(toleranceIsland / areaSize), 1: int(toleranceVoid / areaSize)}
	FullFillMatrix(&areaMap, tolerance, areaSize, adjType, true)
	return WriteTifSquare(flattenAreaMap(areaMap), gdal, OrderedPair{0, 0}, OrderedPair{len(areaMap), len(areaMap[0])}, OrderedPair{len(areaMap), len(areaMap[0])}, outfile, true)

}

func getMatrixFull(filepath string) ([][]Square, GDalInfo, error) {
	flattenedMap, gdal, rowsAndCols, err := ReadTif(filepath)
	if err != nil {
		return [][]Square{}, GDalInfo{}, err
	}

	return makeAreaMap(flattenedMap, rowsAndCols), gdal, nil

}

func bufferSize(adjType int, tolerance map[byte]int) OrderedPair {
	maxTolerance := Max(tolerance[byte(0)], tolerance[byte(1)])
	switch adjType {
	case 4:
		return OrderedPair{maxTolerance, maxTolerance}

	default: // case 8
		return OrderedPair{maxTolerance * 2, maxTolerance * 2}
	}
}

//Return start of chunk, size of chunk (both Ordered Pairs), valid
func makeChunk(buffer OrderedPair, chunkSize OrderedPair, rowsAndCols OrderedPair, r int, c int) (OrderedPair, OrderedPair, [4]int) {
	startBuffer := OrderedPair{Max(0, r-buffer.R), Max(0, c-buffer.C)}
	endBuffer := OrderedPair{Min(rowsAndCols.R, r+chunkSize.R+buffer.R), Min(rowsAndCols.C, c+chunkSize.C+buffer.C)}
	size := OrderedPair{endBuffer.R - startBuffer.R, endBuffer.C - startBuffer.C}

	newR := Min(buffer.R, r)
	newC := Min(buffer.C, c)
	valid := [4]int{newR, Min(newR+chunkSize.R, size.R), newC, Min(newC+chunkSize.C, size.C)}

	fmt.Print(startBuffer, size, valid)

	return startBuffer, size, valid
}

func ChunkFillMap(filepath string, outfile string, toleranceIsland float64, toleranceVoid float64, chunkSize OrderedPair, adjType int, chanSize int) error {
	gdal, rowsAndCols, err := GetTifInfo(filepath)
	// fmt.Println("initial gdal", gdal)

	areaSize := math.Abs(gdal.XCell * gdal.YCell)
	tolerance := map[byte]int{0: int(toleranceIsland / areaSize), 1: int(toleranceVoid / areaSize)}

	if err != nil {
		return err
	}

	buffer := bufferSize(adjType, tolerance)
	chunkChannel := make(chan chunkFillStruct, chanSize)
	i := 0
	for r := 0; r < rowsAndCols.R; r += chunkSize.R {
		for c := 0; c < rowsAndCols.C; c += chunkSize.C {

			start, size, valid := makeChunk(buffer, chunkSize, rowsAndCols, r, c)

			ChunkFillSolve(filepath, gdal, tolerance, start, size, valid, areaSize, adjType, chunkChannel)

			completedChunk := <-chunkChannel
			bufferSize := OrderedPair{len(completedChunk.AreaMap), len(completedChunk.AreaMap[0])}

			WriteTifSquare(flattenAreaMap(completedChunk.AreaMap), gdal, completedChunk.Offset, rowsAndCols, bufferSize, outfile, i == 0)
			i++
		}
	}

	return nil
}

func ChunkFillSolve(filepath string, gdal GDalInfo, tolerance map[byte]int, offset OrderedPair, size OrderedPair, valid [4]int,
	areaSize float64, adjType int, chunkChannel chan chunkFillStruct) {
	areaMap, err := getMatrixChunk(filepath, offset, size)
	if err != nil {
		log.Fatal(err)
	}

	FullFillMatrix(&areaMap, tolerance, areaSize, adjType, false)
	sliceAreaMap(&areaMap, valid)

	// gdal.XMin += gdal.XCell * float64(start.C+valid[0])
	// gdal.YMin += gdal.YCell * float64(start.R+valid[2])

	// fmt.Println(offset, size, valid)
	// fmt.Println(gdal.XMin, gdal.YMin, start)

	chunkChannel <- chunkFillStruct{areaMap, OrderedPair{offset.R + valid[0], offset.C + valid[2]}}

}

func getMatrixChunk(filepath string, start OrderedPair, size OrderedPair) ([][]Square, error) {
	flattenedMap, err := ReadTifChunk(filepath, start, size)
	if err != nil {
		return [][]Square{}, err
	}

	return makeAreaMap(flattenedMap, size), nil
}
