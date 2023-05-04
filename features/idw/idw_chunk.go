package idw

import (
	"fmt"
	"time"

	"github.com/dewberry/v2r/tools"
	"github.com/dewberry/v2r/tools/processing"

	bunyan "github.com/Dewberry/paul-bunyan"
)

type chunkJob struct {
	locs  *map[tools.OrderedPair]tools.Point
	xInfo tools.Info
	yInfo tools.Info
	start tools.OrderedPair
	end   tools.OrderedPair
	exp   float64
}

type chunkIDW struct {
	Pair tools.OrderedPair
	Data [][]float64
}

func makeGridIDW(jobs chan chunkJob, channel chan chunkIDW) {
	for j := range jobs {
		rStart, cStart := tools.PairToRC(j.start)
		rEnd, cEnd := tools.PairToRC(j.end)
		grid := make([][]float64, rEnd-rStart)
		for r := rStart; r < rEnd; r++ {
			grid[r-rStart] = make([]float64, cEnd-cStart)
			for c := cStart; c < cEnd; c++ {
				grid[r-rStart][c-cStart] = calculateIDW(j.locs, j.xInfo, j.yInfo, j.exp, r, c).Weight
			}
		}

		channel <- chunkIDW{j.start, grid}
	}
}

func ChunkSolve(data *map[tools.OrderedPair]tools.Point, outfile string, xInfo tools.Info, yInfo tools.Info, chunkR int, chunkC int, proj string, pow float64, channel chan string) error {
	start := time.Now()

	numRows, numCols := tools.GetDimensions(xInfo, yInfo)

	if chunkR > numRows || chunkC > numCols {
		bunyan.Warnf("chunk size, %v too large for total image %v", tools.MakePair(chunkR, chunkC), tools.MakePair(numRows, numCols))
		bunyan.Warn("Chunk sizes changed to ~1/16 of total size")
		chunkR = numRows / 4
		chunkC = numCols / 4
	}

	bunyan.Debugf("XINFO: %v     YINFO: %v", xInfo, yInfo)
	bunyan.Infof("RowsXCols: [%v X %v]", numRows, numCols)
	bunyan.Infof("Chunk Dim: [%v X %v]", chunkR, chunkC)

	totalChunks := tools.RoundUp(numRows, chunkR) * tools.RoundUp(numCols, chunkC)
	chunkChannel := make(chan chunkIDW, totalChunks)
	jobs := make(chan chunkJob, totalChunks)
	bunyan.Infof("total chunks: %v", totalChunks)
	numWorkers := tools.Min(totalChunks, getChannelSize(chunkR*chunkC))

	bunyan.Infof("buffered channel size: %v", numWorkers)
	for i := 0; i < numWorkers; i++ {
		go makeGridIDW(jobs, chunkChannel)
	}
	for r := 0; r < numRows; r += chunkR {
		for c := 0; c < numCols; c += chunkC {
			job := chunkJob{data, xInfo, yInfo, tools.RCToPair(r, c), tools.RCToPair(tools.Min(r+chunkR, numRows), tools.Min(c+chunkC, numCols)), pow}
			jobs <- job
		}
	}
	close(jobs)
	totalWait := time.Duration(0)
	totalPrint := time.Duration(0)
	progress := tools.Max(1, totalChunks/10)

	totalSize := tools.RCToPair(numRows, numCols)
	gdal := processing.CreateGDalInfo(xInfo.Min, yInfo.Min, xInfo.Step, yInfo.Step, 7, proj)
	for i := 0; i < totalChunks; i++ {
		start := time.Now()
		chunk := <-chunkChannel
		received := time.Now()

		err := writeGDAL(chunk, outfile, gdal, totalSize, i == 0)
		if err != nil {
			return err
		}

		if (i+1)%progress == 0 {
			bunyan.Infof("~%d%%, %v / %v", 100*(i+1)/totalChunks, i+1, totalChunks)
		} else {
			bunyan.Debugf("%v / %v     wait time: % v      print time: %v", i+1, totalChunks, received.Sub(start), time.Since(received))
		}
		totalWait += received.Sub(start)
		totalPrint += time.Since(received)
	}
	bunyan.Debugf("Total Wait: %v", totalWait)
	bunyan.Debugf("Total Print: %v", totalPrint)

	channel <- fmt.Sprintf("pow%v [%vX%v] completed in %v", pow, numRows, numCols, time.Since(start))
	return nil
}
