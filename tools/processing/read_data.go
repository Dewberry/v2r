package processing

import (
	"app/tools"
	"bufio"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
)

func ReadData(filepath string) ([]tools.Point, tools.Info, tools.Info, error) {
	file, err := os.Open(filepath)

	if err != nil {
		return []tools.Point{}, tools.Info{}, tools.Info{}, err
	}
	defer file.Close()

	var data []tools.Point
	xInfo, yInfo := tools.MakeInfo(), tools.MakeInfo()
	sc := bufio.NewScanner(file)
	for sc.Scan() {
		switch strings.Fields(sc.Text())[0] {
		case "POINTS":
			data = addPoints(sc)

		case "STEP":
			sc.Scan()
			fields := strings.Fields(sc.Text())
			xInfo.Step, err = strconv.ParseFloat(strings.TrimSpace(fields[0]), 64)
			if err != nil {
				return []tools.Point{}, tools.Info{}, tools.Info{}, err
			}
			yInfo.Step, err = strconv.ParseFloat(strings.TrimSpace(fields[1]), 64)
			if err != nil {
				return []tools.Point{}, tools.Info{}, tools.Info{}, err
			}

		case "ESTIMATE":
			for xy := 0; xy < 2; xy++ {
				sc.Scan()
				for minMax, val := range strings.Fields(sc.Text()) {
					val, innerErr := strconv.ParseFloat(val, 64)
					if innerErr != nil {
						return []tools.Point{}, tools.Info{}, tools.Info{}, innerErr
					}

					switch tools.MakePair(xy, minMax) {
					case tools.MakePair(0, 0):
						xInfo.Min = val

					case tools.MakePair(0, 1):
						xInfo.Max = val

					case tools.MakePair(1, 0):
						yInfo.Min = val

					case tools.MakePair(1, 1):
						yInfo.Max = val
					}
				}
			}
			fmt.Println("reading in", "X:", xInfo, "\ty:", yInfo)
			return data, xInfo, yInfo, nil

		}

	}
	return data, xInfo, yInfo, errors.New("ESTIMATE not in file")
}

func addPoints(sc *bufio.Scanner) []tools.Point {
	line := sc.Text()
	numPoints, _ := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, "POINTS ")))

	var data []tools.Point
	for i := 0; i < numPoints; i++ {
		sc.Scan()
		var p tools.Point

		fields := strings.Fields(sc.Text())

		p.X, _ = strconv.ParseFloat(fields[0], 64)
		p.Y, _ = strconv.ParseFloat(fields[1], 64)
		p.Weight, _ = strconv.ParseFloat(fields[2], 64)

		fmt.Println(p)

		data = append(data, p)

	}
	return data
}

// Reads Database into a list of points
// Stores min and max x, y values
// Configures globals (min/max x, y)
func ReadPGData(db *sqlx.DB, query string, stepX float64, stepY float64) ([]tools.Point, tools.Info, tools.Info, error) {
	rows, err := db.Query(query)
	if err != nil {
		return []tools.Point{}, tools.Info{}, tools.Info{}, err
	}

	var listPoints []tools.Point
	xInfo, yInfo := tools.MakeInfo(), tools.MakeInfo()

	xInfo.Step = stepX
	yInfo.Step = stepY

	for rows.Next() {
		var elev, x, y float64

		err = rows.Scan(&elev, &x, &y)
		if err != nil {
			return []tools.Point{}, tools.Info{}, tools.Info{}, err
		}

		xInfo.Min = math.Min(xInfo.Min, x)
		xInfo.Max = math.Max(xInfo.Max, x)

		yInfo.Min = math.Min(yInfo.Min, y)
		yInfo.Max = math.Max(yInfo.Max, y)

		listPoints = append(listPoints, tools.MakePoint(x, y, elev))
	}

	return listPoints, xInfo, yInfo, nil
}
