package main

import (
	"app/tools"
	"fmt"
	"log"
)

func main() {
	// From txt file
	// inputFile := "/app/data/small/nb2.txt"
	// err := tools.ReadData(inputFile)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	db := tools.DBInit()

	err := tools.PingWithTimeout(db)
	if err != nil {
		fmt.Println("Connected to database?", err)
	}

	inputQuery := ""
	result, err := tools.ReadPGData(db, inputQuery)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)

}
