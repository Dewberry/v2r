package tools

import (
	"bufio"
	"errors"
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
	X, Y int
}

type Coord struct {
	P    Point
	Pair OrderedPair
}

// func PointToPair(p Point) OrderedPair {
// 	newX := p.X -
// 	return OrderedPair{}
// }

func euclid_dist(p1, p2 Point) float64 {
	total := math.Pow(p1.X-p2.X, 2) + math.Pow(p1.Y-p2.Y, 2)
	return math.Pow(total, .5)
}

func GetWeight(p1 Point) float64 {
	return p1.Weight
}

// p0 is grid location, p is weighted point to compare to
func PartialWeight(p0, p Point, exp float64) float64 {
	return p.Weight / (math.Pow(euclid_dist(p, p0), exp))
}

func DistExp(p0, p Point, exp float64) float64 {
	return math.Pow(euclid_dist(p, p0), -exp)
}

func SetMinMax(min_x, max_x, min_y, max_y, xStep, yStep float64) {
	xSpace := int((max_x - min_x) / xStep)
	ySpace := int((max_y - min_y) / yStep)

	setMinMaxHelper(0, xSpace, 0, ySpace)
}

func setMinMaxHelper(min_x, max_x, min_y, max_y int) {
	MIN[0] = min_x
	MAX[0] = max_x
	MAX[1] = max_y
	MIN[1] = min_y
}

// TODO: implement boundaries
// TODO: implement speedups
func CalculateWeight(cell Point, data []Point, exp float64) float64 {
	total := 0.0
	for _, p := range data {
		total += p.Weight / (math.Pow(euclid_dist(cell, p), exp))
	}
	return math.Pow(total, .5)
}

func createCoordPoint(db_elem Point, min_x float64, min_y float64, xStep float64, yStep float64, ch chan Coord) {
	new_x := int((db_elem.X - min_x) / xStep)
	new_y := int((db_elem.Y - min_y) / yStep)

	co := Coord{Point{db_elem.X, db_elem.Y, db_elem.Weight}, OrderedPair{new_x, new_y}}
	fmt.Println(co, xStep, yStep)
	ch <- Coord{Point{db_elem.X, db_elem.Y, db_elem.Weight}, OrderedPair{new_x, new_y}}
}

//TODO: FIX THIS
func MakeCoordSpace(db_items []Point, min_x float64, min_y float64, xStep float64, yStep float64) map[OrderedPair]Point {
	seen := map[OrderedPair]Point{}

	channel := make(chan Coord, len(db_items))
	for _, db_elem := range db_items {
		go createCoordPoint(db_elem, min_x, min_y, xStep, yStep, channel)
	}

	for i := 0; i < len(db_items); i++ {
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

func ReadIn(dim int, f string) ([]Point, error) {
	file, err := os.Open(f)
	if err != nil {
		return []Point{}, err
	}

	defer file.Close()

	sc := bufio.NewScanner(file)

	var data []Point
	// var bounds []Point // to be implemented later

	for sc.Scan() {
		switch strings.Fields(sc.Text())[0] {
		case "POINTS":
			data = addPoints(sc)

		case "ESTIMATE":
			for d := 0; d < dim; d++ {
				sc.Scan()
				for i, val := range strings.Fields(sc.Text()) {
					val, innerErr := strconv.Atoi(val)
					if innerErr != nil {
						return data, innerErr
					}

					if i == 0 {
						MIN[d] = val
					} else {
						MAX[d] = val
					}

				}

			}
			return data, nil

		}

	}
	return data, errors.New("ESTIMATE not in file")

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
func PrintExcel(data [][]float64, filename string, sheetname string) error {
	grid := Transpose(data)

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
		file.SetCellValue(sheetname, fmt.Sprintf("A%v", i+1), MAX[1]-i) // y-axis
		file.SetSheetRow(sheetname, fmt.Sprintf("B%v", len(grid)-i), &row)
		file.SetRowHeight(sheetname, i+1, 25)

	}
	file.SetRowHeight(sheetname, endrow+1, 25)

	for i := 0; i < len(grid[0]); i++ {
		file.SetCellValue(sheetname, fmt.Sprintf("%s%v", GetExcelColumn(i+1), len(grid)+1), i+MIN[0]) // x-axis
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
