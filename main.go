package main

import (
	"app/tools"
	"fmt"
	"log"
)

func main() {

	db := tools.DBInit()

	err := tools.PingWithTimeout(db)
	if err != nil {
		fmt.Println("Connected to database?", err)
	}

	// inputFile := "/app/data/example.txt"
	// err = tools.ReadData(inputFile)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	inputQuery := ""
	result, err := tools.ReadPGData(db, inputQuery)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)
}
