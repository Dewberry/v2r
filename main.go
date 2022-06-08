package main

import (
	"app/tools"
	"fmt"
	"log"
)

func main() {
	fmt.Println("R2V Dev")
	inputFile := "/app/data/example.txt"
	err := tools.ReadData(inputFile)
	if err != nil {
		log.Fatal(err)
	}
}
