package idw

import (
	"fmt"
	"strings"
	"time"

	"github.com/dewberry/v2r/tools"
	"github.com/dewberry/v2r/tools/processing"

	bunyan "github.com/Dewberry/paul-bunyan"
)

// Run IDW algorithm serially. Print to outfile
func FullSolve(data *map[tools.OrderedPair]tools.Point, outfile string, xInfo tools.Info, yInfo tools.Info, proj string, pow float64, channel chan string) error {
	start := time.Now()

	// outfile checking
	ascii := strings.HasSuffix(outfile, ".asc")
	excel := strings.HasSuffix(outfile, ".xlsx")
	tiff := strings.HasSuffix(outfile, ".tiff") || strings.HasSuffix(outfile, ".tif")

	if !(ascii || excel || tiff) {
		return fmt.Errorf("outfile is not supported. Use one of these files: tiff (.tif/.tiff), ascii (.asc), or excel (.xlsx).")
	}

	numRows, numCols := tools.GetDimensions(xInfo, yInfo)
	bunyan.Debugf("XINFO: %v     YINFO: %v", xInfo, yInfo)
	bunyan.Debugf("[%v X %v]", numRows, numCols)

	totalSize := tools.RCToPair(numRows, numCols)
	gdal := processing.CreateGDalInfo(xInfo.Min, yInfo.Min, xInfo.Step, yInfo.Step, 7, proj)

	grid := make([][]float64, totalSize.R)
	for r := 0; r < totalSize.R; r++ {
		grid[r] = make([]float64, totalSize.C)
		for c := 0; c < totalSize.C; c++ {
			grid[r][c] = calculateIDW(data, xInfo, yInfo, pow, r, c).Weight
		}
	}

	if excel {
		bunyan.Debug("excel write")
		err := processing.PrintExcel(grid, outfile, pow)
		if err != nil {
			return err
		}
	} else { // ascii or tiff write
		if ascii {
			bunyan.Debug("ascii write")
		} else if tiff {
			bunyan.Debug("tiff write")
		}

		toPrint := chunkIDW{tools.MakePair(0, 0), grid}
		err := writeGDAL(toPrint, outfile, gdal, totalSize, true)

		if err != nil {
			return err
		}
	}

	channel <- fmt.Sprintf("pow%v [%vX%v] completed in %v", pow, numRows, numCols, time.Since(start))
	return nil
}
