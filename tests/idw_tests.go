package tests

import (
	"app/features/idw"
	tools "app/tools"
	processing "app/tools/processing"
	"fmt"
	"log"
)

func testIDW() {
	chunkR := 3
	chunkC := 2
	pow := 1.7
	stepX := 3.0
	stepY := 2.0
	epsg := 2284
	outfileFolder := "tests/"

	// From txt file
	inputFile := "tests/idw_in.txt"
	listPoints, xInfo, yInfo, err := processing.ReadData(inputFile)
	if err != nil {
		log.Fatal(err)
	}
	// From txt file

	data := tools.MakeCoordSpace(&listPoints, xInfo, yInfo)
	outfileFull := fmt.Sprintf("%sstep%.0f-%.0f", outfileFolder, stepX, stepY) // "step{x}-{y}[chunked]pow{pow}.[ext]"
	// outfileChunk := fmt.Sprintf("%schunked", outfileFull)
	channel := make(chan string, 2)

	go idw.MainSolve(&data, outfileFull, xInfo, yInfo, pow, false, chunkR, chunkC, epsg, channel)
	// go idw.MainSolve(&data, outfileChunk, xInfo, yInfo, pow, true, chunkR, chunkC, epsg, channel)

	for i := 0; i < 1; i++ {
		receivedString := <-channel
		fmt.Println(receivedString)
	}

	completeOutfileTif := fmt.Sprintf("%spow1.7.tiff", outfileFull)
	completeOutfileAsc := fmt.Sprintf("%spow1.7.asc", outfileFull)
	fmt.Println("pre transfer")
	processing.TransferType(completeOutfileTif, completeOutfileAsc)
	// completeOutfileChunked := fmt.Sprintf("%spow1.7.tiff", outfileChunk)
	correct := "idw_correct.asc"
	fmt.Printf("File output correct\n____________________________\n")
	fmt.Printf("NO CHUNKING: %v\n", sameFiles(completeOutfileAsc, correct))
	// fmt.Printf("CHUNKING: %v\n", sameFiles(completeOutfileChunked, correctChunked))
}
