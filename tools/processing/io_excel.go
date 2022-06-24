package processing

import (
	"fmt"
	"strings"

	bunyan "github.com/Dewberry/paul-bunyan"
	"github.com/xuri/excelize/v2"
)

func GetExcelColumn(i int) string {
	i++ // 1-indexed
	endcol, err := excelize.CoordinatesToCellName(i, 1)
	if err != nil {
		bunyan.Fatal(err)
	}
	return strings.TrimRight(endcol, "1")

}

//Can not be used concurrently
func PrintExcel(grid [][]float64, filepath string, pow float64) error {
	filename := fmt.Sprintf("%s.xlsx", filepath)
	sheetname := fmt.Sprintf("pow%v", pow)

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
