package main

import (
	"app/tools"
	"fmt"
	"log"
)

func main() {
	useChunk := true

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
	stepX := 20.0
	stepY := 20.0
	result, err := tools.ReadPGData(db, inputQuery, stepX, stepY, useChunk)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result)

}
