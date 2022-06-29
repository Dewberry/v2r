package cleaner

import (
	"github.com/dewberry/v2r/tools"
	"github.com/dewberry/v2r/tools/processing"

	"math"

	"time"

	bunyan "github.com/Dewberry/paul-bunyan"
)

type chunkJob struct {
	Filepath  string
	Tolerance map[byte]int
	Offset    tools.OrderedPair
	ICP       innerChunkPartition
	Size      tools.OrderedPair
	AreaSize  float64
	AdjType   int
}

type chunkFillStruct struct {
	AreaMap [][]square
	cStats  cleanerStats
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

func cleanChunk(jobs chan chunkJob, chunkChannel chan chunkFillStruct) {
	for j := range jobs {
		areaMap, err := readFileChunk(j.Filepath, tools.MakePair(j.Offset.R-j.ICP.RStart, j.Offset.C-j.ICP.CStart), j.Size)
		if err != nil {
			bunyan.Fatal(err)
		}

		cStats := cleanAreaMap(&areaMap, j.Tolerance, j.AreaSize, j.AdjType, j.ICP)
		sliceAreaMap(&areaMap, j.ICP)

		chunkChannel <- chunkFillStruct{areaMap, cStats, j.Offset}
	}

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

func CleanWithChunking(filepath string, outfile string, toleranceIsland float64, toleranceVoid float64, chunkSize tools.OrderedPair, adjType int) error {
	gdal, rowsAndCols, err := processing.GetInfoGDAL(filepath)
	bunyan.Infof("img rows x cols: %v", rowsAndCols)
	if err != nil {
		bunyan.Fatal(err)
	}

	areaSize := math.Abs(gdal.XCell * gdal.YCell)
	tolerance := map[byte]int{0: int(toleranceIsland / areaSize), 1: int(toleranceVoid / areaSize)}

	buffer := bufferSize(adjType, tolerance)

	cStats := cleanerStats{}
	if chunkSize.R > rowsAndCols.R || chunkSize.C > rowsAndCols.C {
		bunyan.Warnf("chunk size, %v too large for total image %v", chunkSize, rowsAndCols)
		bunyan.Warn("Chunk sizes changed to ~1/16 of total size")
		chunkSize.R = rowsAndCols.R / 4
		chunkSize.C = rowsAndCols.C / 4
	}
	totalChunks := tools.RoundUp(rowsAndCols.R, chunkSize.R) * tools.RoundUp(rowsAndCols.C, chunkSize.C)

	chunkChannel := make(chan chunkFillStruct, totalChunks)
	jobs := make(chan chunkJob, totalChunks)
	bunyan.Infof("total chunks: %v", totalChunks)

	numWorkers := tools.Min(totalChunks, getChannelSize(chunkSize.R*chunkSize.C))
	bunyan.Infof("buffered channel size: %v", numWorkers)
	for i := 0; i < numWorkers; i++ {
		go cleanChunk(jobs, chunkChannel)
	}
	for r := 0; r < rowsAndCols.R; r += chunkSize.R {
		for c := 0; c < rowsAndCols.C; c += chunkSize.C {
			innerChunk, size := makeChunk(buffer, chunkSize, rowsAndCols, r, c)

			jobs <- chunkJob{filepath, tolerance, tools.MakePair(r, c), innerChunk, size, areaSize, adjType}
		}
	}
	close(jobs)
	totalWait := time.Duration(0)
	totalPrint := time.Duration(0)
	progress := totalChunks / 10
	for j := 0; j < totalChunks; j++ {
		start := time.Now()
		completedChunk := <-chunkChannel
		received := time.Now()

		bufferSize := tools.MakePair(len(completedChunk.AreaMap), len(completedChunk.AreaMap[0]))
		cStats.updateStats(completedChunk.cStats)

		err = processing.WriteTif(flattenAreaMap(completedChunk.AreaMap), gdal, outfile, completedChunk.Offset, rowsAndCols, bufferSize, j == 0)
		if err != nil {
			bunyan.Fatal(err)
		}
		if (j+1)%progress == 0 {
			bunyan.Infof("~%d%%, %v / %v", 100*(j+1)/totalChunks, j+1, totalChunks)
		} else {
			bunyan.Debugf("%v / %v     wait time: % v      print time: %v", j+1, totalChunks, received.Sub(start), time.Since(received))
		}
		totalWait += received.Sub(start)
		totalPrint += time.Since(received)
	}
	bunyan.Debugf("Total Wait:  %v", totalWait)
	bunyan.Debugf("Total Print: %v", totalPrint)
	printStats(cStats, areaSize)

	return nil
}
