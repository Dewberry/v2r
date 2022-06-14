package tools

import (
	"bufio"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
)

func ReadData(filepath string) ([]Point, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return []Point{}, err
	}

	defer file.Close()

	sc := bufio.NewScanner(file)

	var data []Point
	// var bounds []Point // to be implemented later

	stepX, stepY := 1.0, 1.0
	for sc.Scan() {
		switch strings.Fields(sc.Text())[0] {
		case "POINTS":
			data = addPoints(sc)

		case "STEP":
			sc.Scan()
			fields := strings.Fields(sc.Text())
			stepX, err = strconv.ParseFloat(strings.TrimSpace(fields[0]), 64)
			if err != nil {
				return []Point{}, err
			}
			stepY, err = strconv.ParseFloat(strings.TrimSpace(fields[1]), 64)
			if err != nil {
				return []Point{}, err
			}

		case "ESTIMATE":
			for d := 0; d < 2; d++ {
				sc.Scan()
				for i, val := range strings.Fields(sc.Text()) {
					val, innerErr := strconv.ParseFloat(val, 64)
					if innerErr != nil {
						return data, innerErr
					}

					if d == 0 {
						GlobalX[i] = val
					} else {
						GlobalY[i] = val
					}

				}
			}
			GlobalX[2] = stepX
			GlobalY[2] = stepY
			fmt.Println("reading in", "X:", GlobalX, "\ty:", GlobalY)
			return data, nil

		}

	}
	return data, errors.New("ESTIMATE not in file")
}

// Reads Database into a list of points
// Stores min and max x, y values
// Configures globals (min/max x, y)
func ReadPGData(db *sqlx.DB, query string, stepX float64, stepY float64) ([]Point, error) {
	rows, err := db.Query(query)
	if err != nil {
		return []Point{}, err
	}

	var listPoints []Point
	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)

	for rows.Next() {
		var elev, x, y float64

		err = rows.Scan(&elev, &x, &y)
		if err != nil {
			return []Point{}, err
		}

		minX = math.Min(minX, x)
		maxX = math.Max(maxX, x)

		minY = math.Min(minY, y)
		maxY = math.Max(maxY, y)

		listPoints = append(listPoints, Point{x, y, elev})
	}
	ConfigureGlobals(minX, maxX, stepX, minY, maxY, stepY)

	return listPoints, nil
}
