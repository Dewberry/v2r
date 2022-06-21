package idw

import (
	"app/tools"
	processing "app/tools/processing"
	"fmt"
	_ "net/http/pprof"
	"time"
)

func MainSolve(data *map[tools.OrderedPair]tools.Point, outfile string, xInfo tools.Info, yInfo tools.Info, pow float64, useChunking bool, chunkR int, chunkC int, epsg int, channel chan string) error {
	start := time.Now()

	numRows, numCols := tools.GetDimensions(xInfo, yInfo)
	fmt.Printf("XINFO: %v\nYINFO: %v\n", xInfo, yInfo)
	fmt.Printf("[%v X %v]\n", numRows, numCols)

	if !useChunking {
		chunkR = numRows
		chunkC = numCols
	}
	chunkChannel := make(chan chunkIDW, 1000)
	totalChunks := chunkSolve(data, xInfo, yInfo, pow, chunkR, chunkC, chunkChannel)

	totalSize := tools.RCToPair(numRows, numCols)
	gdal := processing.CreateGDalInfo(xInfo.Min, yInfo.Min, xInfo.Step, yInfo.Step, 7, epsg)
	for i := 0; i < totalChunks; i++ {
		chunk := <-chunkChannel
		writeTif(chunk, fmt.Sprintf("%spow%.1f", outfile, pow), gdal, totalSize, i)
		writeAsc(chunk, fmt.Sprintf("%spow%.1f", outfile, pow), gdal, totalSize, i)
		if !useChunking {
			processing.PrintExcel(chunk.Data, outfile, pow)
		}

	}
	fmt.Printf("chunk sizes: [%v, %v]\ttotal chunks: %v\n", chunkR, chunkC, totalChunks)

	// updateChannel := make(chan bool, 100)
	// for i := 0; i < totalChunks; i++ {
	// 	gridChunk := <-chunkChannel
	// 	go chunkUpdate(&grid, &gridChunk, updateChannel)
	// }

	// for i := 0; i < totalChunks; i++ {
	// 	<-updateChannel
	// }

	// if !useChunking {
	// 	// innerErr := PrintExcel(gridChunk, fmt.Sprintf("%spow%.1f", outfile, pow), pow)
	// 	// innerErr := PrintAscii(gridChunk, fmt.Sprintf("%spow%.1f", outfile, pow), pow, xInfo.Step, yInfo.Step)

	// 	if innerErr != nil {
	// 		return innerErr
	// 	}
	// }
	channel <- fmt.Sprintf("pow%v [%vX%v] completed in %v", pow, numRows, numCols, time.Since(start))
	return nil
}
