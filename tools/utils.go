package tools

import (
	"bufio"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

type Point struct {
	X, Y   int
	Weight float64
}

func euclid_dist(p1, p2 Point) float64 {
	total := math.Pow(float64(p1.X)-float64(p2.X), 2) + math.Pow(float64(p1.Y)-float64(p2.Y), 2)
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

// TODO: implement boundaries
// TODO: implement speedups
func CalculateWeight(cell Point, data []Point, exp float64) float64 {
	total := 0.0
	for _, p := range data {
		total += p.Weight / (math.Pow(euclid_dist(cell, p), exp))
	}
	return math.Pow(total, .5)
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
			for i := 0; i < dim; i++ {
				sc.Scan()
				for i, val := range strings.Fields(sc.Text()) {
					val, innerErr := strconv.Atoi(val)
					if innerErr != nil {
						return data, innerErr
					}

					if i == 0 {
						MIN = append(MIN, val)
					} else {
						MAX = append(MAX, val)
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

		p.X, _ = strconv.Atoi(fields[0])
		p.Y, _ = strconv.Atoi(fields[1])
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

	endcol := rune(65 + len(grid[0]))
	endrow := len(grid)

	for i, row := range grid {
		file.SetCellValue(sheetname, fmt.Sprintf("A%v", i+1), MAX[1]-i) // y-axis
		file.SetSheetRow(sheetname, fmt.Sprintf("B%v", len(grid)-i), &row)
		file.SetRowHeight(sheetname, i+1, 25)

	}
	file.SetRowHeight(sheetname, endrow+1, 25)

	for i := 0; i < len(grid[0]); i++ {
		file.SetCellValue(sheetname, fmt.Sprintf("%c%v", rune(66+i), len(grid)+1), i+MIN[0]) // x-axis
	}
	file.SetColWidth(sheetname, "A", fmt.Sprintf("%c", endcol), 5)

	style, err := file.NewStyle(&excelize.Style{DecimalPlaces: 1})

	if err != nil {
		return err
	}
	err = file.SetCellStyle(sheetname, "A1", fmt.Sprintf("%c%v", endcol, endrow), style)
	if err != nil {
		return err
	}

	file.SetConditionalFormat(sheetname, fmt.Sprintf("B1:%c%v", rune(65+len(grid[0])), len(grid)), `[
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
