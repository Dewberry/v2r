package main

import (
	"app/tools"
	"fmt"
	"log"
)

func main() {
	useChunk := false

	// From txt file
	// inputFile := "data/small/nb2.txt"
	// result, err := tools.ReadData(inputFile, useChunk)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	db := tools.DBInit()

	err := tools.PingWithTimeout(db)
	if err != nil {
		fmt.Println("Connected to database?", err)
	}

	inputQuery := ""
	//step must be the same for qgis output
	xStep := 100.0
	yStep := 100.0
	result, err := tools.ReadPGData(db, inputQuery, xStep, yStep, useChunk)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result)

}
