package processing

import (
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

func GetExcelColumn(i int) (string, error) {
	i++ // 1-indexed
	endcol, err := excelize.CoordinatesToCellName(i, 1)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(endcol, "1"), nil

}

// Print out grid to excel file. Can only be done serially.
func PrintExcel(grid [][]float64, filename string, pow float64) error {
	sheetname := fmt.Sprintf("pow%v", pow)

	file, err := excelize.OpenFile(filename)
	if err != nil {
		file = excelize.NewFile()
		file.SetSheetName("Sheet1", sheetname)
	} else {
		file.DeleteSheet(sheetname)
		file.NewSheet(sheetname)
	}

	endcol, err := GetExcelColumn(len(grid[0]))
	if err != nil {
		return err
	}

	endrow := len(grid)

	for i, row := range grid {
		file.SetCellValue(sheetname, fmt.Sprintf("A%v", i+1), len(grid)-i-1) // y-axis
		file.SetSheetRow(sheetname, fmt.Sprintf("B%v", len(grid)-i), &row)
		file.SetRowHeight(sheetname, i+1, 25)

	}
	file.SetRowHeight(sheetname, endrow+1, 25)

	for i := 0; i < len(grid[0]); i++ {
		col, err := GetExcelColumn(i + 1)
		if err != nil {
			return err
		}
		file.SetCellValue(sheetname, fmt.Sprintf("%s%v", col, len(grid)+1), i) // x-axis
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

	col, err := GetExcelColumn(len(grid[0]))
	if err != nil {
		return err
	}
	file.SetConditionalFormat(sheetname, fmt.Sprintf("B1:%s%v", col, len(grid)), `[
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
