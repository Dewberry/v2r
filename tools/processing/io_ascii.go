package processing

import (
	"app/tools"
	"bufio"
	"fmt"
	"log"
	"os"
)

func PrintAscii(grid [][]float64, filepath string, xInfo tools.Info, yInfo tools.Info, pow float64, chunkR int, chunkC int) error {
	filename := fmt.Sprintf("%s.asc", filepath)

	f, err := os.Create(filename)
	if err != nil {
		return err
	}

	numRows, numCols := tools.GetDimensions(xInfo, yInfo)
	writer := bufio.NewWriter(f)
	header := []string{
		fmt.Sprintf("ncols\t%v", numCols),
		fmt.Sprintf("\nnrows\t%v", numRows),
		fmt.Sprintf("\nyllcorner\t%.1f", yInfo.Min),
		fmt.Sprintf("\nxllcorner\t%.1f", xInfo.Min),
		fmt.Sprintf("\ncellsize\t%.1f", xInfo.Step), // assumed square
		"\nNODATA_value\t-9999",
	}

	for _, line := range header {
		_, err := writer.WriteString(line)
		if err != nil {
			log.Fatalf("Got error while writing to a file. Err: %s", err.Error())
		}
	}

	stringChannel := make(chan StringInt, numRows)
	for r := 0; r < numRows; r++ { // without geotransform, do r = numRows - 1; r >= 0; r-- for the three loops
		go makeString(grid, r, xInfo, yInfo, stringChannel)
	}
	rows := make([]string, numRows)
	for r := 0; r < numRows; r++ {
		stringInt := <-stringChannel
		rows[stringInt.Row] = stringInt.PrintString
	}
	for r := 0; r < numRows; r++ {
		writer.WriteString(rows[r])
		writer.Flush()
	}

	return nil
}

type StringInt struct {
	PrintString string
	Row         int
}

func makeString(grid [][]float64, currentRow int, xInfo tools.Info, yInfo tools.Info, stringChannel chan StringInt) {
	_, numCols := tools.GetDimensions(xInfo, yInfo)

	outstring := "\n"
	for c := 0; c < numCols; c++ {
		outstring += fmt.Sprintf("%.2f ", grid[currentRow][c])
	}
	stringChannel <- StringInt{outstring, currentRow}
}
