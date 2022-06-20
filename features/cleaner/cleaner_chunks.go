package cleaner

import (
	"app/tools"
	processing "app/tools/processing"
	"fmt"

	"log"
	"math"
)

type chunkFillStruct struct {
	AreaMap [][]square
	Offset  tools.OrderedPair
}

type innerChunkPartition struct {
	RStart int
	REnd   int
	CStart int
	CEnd   int
}

func sliceAreaMap(areaMap *[][]square, ICP innerChunkPartition) {
	*areaMap = (*areaMap)[ICP.RStart:ICP.REnd]
	for r := 0; r < len(*areaMap); r++ {
		(*areaMap)[r] = (*areaMap)[r][ICP.CStart:ICP.CEnd]
	}
}

func cleanChunk(filepath string, tolerance map[byte]int, offset tools.OrderedPair, ICP innerChunkPartition,
	size tools.OrderedPair, areaSize float64, adjType int, chunkChannel chan chunkFillStruct) {

	areaMap, err := readFileChunk(filepath, tools.MakePair(offset.R-ICP.RStart, offset.C-ICP.CStart), size)
	if err != nil {
		log.Fatal(err)
	}

	cleanAreaMap(&areaMap, tolerance, areaSize, adjType, ICP, false)
	sliceAreaMap(&areaMap, ICP)

	chunkChannel <- chunkFillStruct{areaMap, offset}
}

func bufferSize(adjType int, tolerance map[byte]int) tools.OrderedPair {
	maxTolerance := tools.Max(tolerance[byte(0)], tolerance[byte(1)])
	switch adjType {
	case 4:
		return tools.MakePair(maxTolerance, maxTolerance)

	default: // case 8
		return tools.MakePair(maxTolerance*2, maxTolerance*2)
	}
}

//Return start of chunk, size of chunk (both Ordered Pairs), valid
func makeChunk(buffer tools.OrderedPair, chunkSize tools.OrderedPair, rowsAndCols tools.OrderedPair, r int, c int) (innerChunkPartition, tools.OrderedPair) {
	startBuffer := tools.MakePair(tools.Max(0, r-buffer.R), tools.Max(0, c-buffer.C))
	endBuffer := tools.MakePair(tools.Min(rowsAndCols.R, r+chunkSize.R+buffer.R), tools.Min(rowsAndCols.C, c+chunkSize.C+buffer.C))
	size := tools.MakePair(endBuffer.R-startBuffer.R, endBuffer.C-startBuffer.C)

	newR := tools.Min(buffer.R, r)
	newC := tools.Min(buffer.C, c)
	innerChunk := innerChunkPartition{newR, tools.Min(newR+chunkSize.R, size.R), newC, tools.Min(newC+chunkSize.C, size.C)}

	return innerChunk, size
}

func CleanWithChunking(filepath string, outfile string, toleranceIsland float64, toleranceVoid float64, chunkSize tools.OrderedPair, adjType int, chanSize int) error {
	gdal, rowsAndCols, err := processing.GetTifInfo(filepath)
	if err != nil {
		return err
	}

	areaSize := math.Abs(gdal.XCell * gdal.YCell)
	tolerance := map[byte]int{0: int(toleranceIsland / areaSize), 1: int(toleranceVoid / areaSize)}

	buffer := bufferSize(adjType, tolerance)
	chunkChannel := make(chan chunkFillStruct, chanSize)
	i := 0
	for r := 0; r < rowsAndCols.R; r += chunkSize.R {
		for c := 0; c < rowsAndCols.C; c += chunkSize.C {
			fmt.Println(i)

			innerChunk, size := makeChunk(buffer, chunkSize, rowsAndCols, r, c)

			cleanChunk(filepath, tolerance, tools.MakePair(r, c), innerChunk, size, areaSize, adjType, chunkChannel)

			completedChunk := <-chunkChannel
			bufferSize := tools.MakePair(len(completedChunk.AreaMap), len(completedChunk.AreaMap[0]))

			processing.WriteTif(flattenAreaMap(completedChunk.AreaMap), gdal, outfile, completedChunk.Offset, rowsAndCols, bufferSize, i == 0)
			i++
		}
	}

	return nil
}
