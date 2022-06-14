package tools

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

type Point struct {
	X, Y   float64
	Weight float64
}

type OrderedPair struct {
	R, C int
}

type Coord struct {
	P    Point
	Pair OrderedPair
}

func PointToPair(p Point) OrderedPair {
	return OrderedPair{int((p.Y - GlobalY[0]) / GlobalY[2]), int((p.X - GlobalX[0]) / GlobalX[2])}
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func RCToPoint(r, c int) Point {
	px := GlobalX[0] + float64(c)*GlobalX[2]
	py := GlobalY[0] + float64(r)*GlobalY[2]
	return (Point{px, py, 0})
}
func PairToPoint(pair OrderedPair) Point {
	return RCToPoint(pair.R, pair.C)
}

func RCToPair(r, c int) OrderedPair {
	return OrderedPair{r, c}
}

func PairToRC(pair OrderedPair) (int, int) {
	return pair.R, pair.C
}

func GetDimensions() (int, int) {
	cols := int((1 + GlobalX[1] - GlobalX[0]) / GlobalX[2]) // (1 + max - min)/step
	rows := int((1 + GlobalY[1] - GlobalY[0]) / GlobalY[2])
	return rows, cols
}

func getChunkBlock(row, chunkR, chunkC int) (int, int) {
	_, numCols := GetDimensions()
	// numChunksOnRow := int(math.Ceil(float64(numRows) / float64(chunkR)))
	numChunksOnCol := int(math.Ceil(float64(numCols) / float64(chunkC)))

	start := row / chunkR * numChunksOnCol
	return start, start + numChunksOnCol

}

func euclidDist(p1, p2 Point) float64 {
	total := math.Pow(p1.X-p2.X, 2) + math.Pow(p1.Y-p2.Y, 2)
	return math.Pow(total, .5)
}

func GetWeight(p1 Point) float64 {
	return p1.Weight
}

// p0 is grid location, p is weighted point to compare to
func PartialWeight(p0, p Point, exp float64) float64 {
	return p.Weight / (math.Pow(euclidDist(p, p0), exp))
}

func DistExp(p0, p Point, exp float64) float64 {
	return math.Pow(euclidDist(p, p0), -exp)
}

// TODO: implement boundaries
// TODO: implement speedups
func CalculateWeight(cell Point, data *[]Point, exp float64) float64 {
	total := 0.0
	for _, p := range *data {
		total += p.Weight / (math.Pow(euclidDist(cell, p), exp))
	}
	return math.Pow(total, .5)
}

func createCoordPoint(p Point, ch chan Coord) {
	pair := PointToPair(p)
	ch <- Coord{Point{p.X, p.Y, p.Weight}, pair}
}

func MakeCoordSpace(listPoints *[]Point) map[OrderedPair]Point {
	seen := map[OrderedPair]Point{}

	channel := make(chan Coord, len(*listPoints))
	for _, p := range *listPoints {
		go createCoordPoint(p, channel)
	}

	for i := 0; i < len(*listPoints); i++ {
		dataPoint := <-channel

		pair := dataPoint.Pair
		p := dataPoint.P
		elev, exists := seen[pair]
		if exists {
			newElev := (p.Weight + elev.Weight) / 2
			fmt.Printf("%v already exists\nold elev: %v\t this elev%v\nave elev: %v\n______\n", pair, elev, p.Weight, newElev)
			p.Weight = newElev
		}
		seen[pair] = p
	}
	return seen
}

func addPoints(sc *bufio.Scanner) []Point {
	line := sc.Text()
	numPoints, _ := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, "POINTS ")))

	var data []Point
	for i := 0; i < numPoints; i++ {
		sc.Scan()
		var p Point

		fields := strings.Fields(sc.Text())

		p.X, _ = strconv.ParseFloat(fields[0], 64)
		p.Y, _ = strconv.ParseFloat(fields[1], 64)
		p.Weight, _ = strconv.ParseFloat(fields[2], 64)

		fmt.Println(p)

		data = append(data, p)

	}
	return data

}

func Transpose(a [][]float64) [][]float64 {
	m := len(a[0])
	n := len(a)

	b := make([][]float64, m)
	for i := 0; i < m; i++ {
		b[i] = make([]float64, n)
		for j := 0; j < n; j++ {
			b[i][j] = a[j][i]
		}
	}

	return b
}

func GetExcelColumn(i int) string {
	i++ // 1-indexed
	endcol, err := excelize.CoordinatesToCellName(i, 1)
	if err != nil {
		log.Fatal(err)
	}
	return strings.TrimRight(endcol, "1")

}
func PrintExcel(grid [][]float64, filepath string, pow float64) error {
	filename := fmt.Sprintf("%s.xlsx", filepath)
	sheetname := fmt.Sprintf("pow%v", pow)

	// grid := Transpose(data)

	file, err := excelize.OpenFile(filename)
	if err != nil {
		file = excelize.NewFile()
		file.SetSheetName("Sheet1", sheetname)
	} else {
		file.DeleteSheet(sheetname)
		file.NewSheet(sheetname)
	}

	endcol := GetExcelColumn(len(grid[0]))
	endrow := len(grid)

	for i, row := range grid {
		// go printRowHelper(file, sheetname, fmt.Sprintf("A%v", i+1), fmt.Sprintf("B%v", len(grid)-i), i+1, MAX[1]-i, row, 25)
		file.SetCellValue(sheetname, fmt.Sprintf("A%v", i+1), len(grid)-i-1) // y-axis
		file.SetSheetRow(sheetname, fmt.Sprintf("B%v", len(grid)-i), &row)
		file.SetRowHeight(sheetname, i+1, 25)

	}
	file.SetRowHeight(sheetname, endrow+1, 25)

	for i := 0; i < len(grid[0]); i++ {
		file.SetCellValue(sheetname, fmt.Sprintf("%s%v", GetExcelColumn(i+1), len(grid)+1), i) // x-axis
	}
	file.SetColWidth(sheetname, "A", endcol, 5)

	style, err := file.NewStyle(&excelize.Style{DecimalPlaces: 1})

	if err != nil {
		return err
	}
	err = file.SetCellStyle(sheetname, "A1", fmt.Sprintf("%s%v", endcol, endrow), style)
	if err != nil {
		return err
	}

	file.SetConditionalFormat(sheetname, fmt.Sprintf("B1:%s%v", GetExcelColumn(len(grid[0])), len(grid)), `[
		{
			"type": "3_color_scale",
			"criteria": "=",
			"min_type": "min",
			"mid_type": "percentile",
			"max_type": "max",
			"min_color": "#63BE7B",
			"mid_color": "#FFEB84",
			"max_color": "#F8696B"
		}]`)

	file.SaveAs(filename)

	return nil
}

func PrintAscii(grid [][]float64, filepath string, pow float64, chunkR int, chunkC int) error {
	filename := fmt.Sprintf("%s.asc", filepath)

	f, err := os.Create(filename)
	if err != nil {
		return err
	}

	numRows, numCols := GetDimensions()
	writer := bufio.NewWriter(f)
	header := []string{
		fmt.Sprintf("ncols\t%v", numCols),
		fmt.Sprintf("\nnrows\t%v", numRows),
		fmt.Sprintf("\nyllcorner\t%.1f", GlobalY[0]),
		fmt.Sprintf("\nxllcorner\t%.1f", GlobalX[0]),
		fmt.Sprintf("\ncellsize\t%.1f", CELL),
		"\nNODATA_value\t-9999",
	}

	for _, line := range header {
		_, err := writer.WriteString(line)
		if err != nil {
			log.Fatalf("Got error while writing to a file. Err: %s", err.Error())
		}
	}

	stringChannel := make(chan StringInt, numRows)
	for r := numRows - 1; r >= 0; r-- {
		go makeString(grid, r, stringChannel)
	}
	rows := make([]string, numRows)
	for r := numRows - 1; r >= 0; r-- {
		stringInt := <-stringChannel
		rows[stringInt.Row] = stringInt.PrintString
	}
	for r := numRows - 1; r >= 0; r-- {
		writer.WriteString(rows[r])
		writer.Flush()
	}

	return nil
}

type StringInt struct {
	PrintString string
	Row         int
}

func makeString(grid [][]float64, currentRow int, stringChannel chan StringInt) {
	_, numCols := GetDimensions()

	outstring := "\n"
	for c := 0; c < numCols; c++ {
		outstring += fmt.Sprintf("%.2f ", grid[currentRow][c])
	}
	stringChannel <- StringInt{outstring, currentRow}
}
