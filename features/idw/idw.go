package idw

import (
	"fmt"
	"time"

	"github.com/dewberry/v2r/tools"
	"github.com/dewberry/v2r/tools/processing"

	bunyan "github.com/Dewberry/paul-bunyan"
)

func FullSolve(data *map[tools.OrderedPair]tools.Point, outfile string, xInfo tools.Info, yInfo tools.Info, proj string, pow float64, ascii bool, excel bool, channel chan string) error {
	start := time.Now()

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
	toPrint := chunkIDW{tools.MakePair(0, 0), grid}
	outfile = fmt.Sprintf("%spow%.1f", outfile, pow)
	err := writeTif(toPrint, outfile, gdal, totalSize, 0)
	if err != nil {
		return err
	}
	if ascii {
		bunyan.Debug("ascii write")
		err := processing.TransferType(outfile+".tiff", outfile+".asc", "Int32") // for ascii output
		if err != nil {
			return err
		}
	}
	if excel {
		bunyan.Debug("excel write")
		err := processing.PrintExcel(grid, outfile, pow)
		if err != nil {
			return err
		}
	}

	channel <- fmt.Sprintf("pow%v [%vX%v] completed in %v", pow, numRows, numCols, time.Since(start))
	return nil
}
