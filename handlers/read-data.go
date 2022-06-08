package handlers

import (
	"app/tools"
	"net/http"

	// "github.com/v2r/tools"

	"strconv"

	"github.com/labstack/echo/v4"
)

func ReadData() echo.HandlerFunc {
	return func(c echo.Context) error {

		filepath := c.QueryParam("filepath")

		print_out, innerErr := strconv.ParseBool(c.QueryParam("print_out"))
		if innerErr != nil {
			return innerErr
		}

		data, innerErr := tools.ReadIn(2, filepath) // 2 dimensions hardcoded
		if innerErr != nil {
			return innerErr
		}

		for pow := 1.0; pow < 3.5; pow += .5 {
			innerErr = tools.MainSolve(data, filepath, pow, print_out)
			if innerErr != nil {
				return innerErr
			}
			// return nil
		}

		return c.JSON(http.StatusOK, "read path correctly")
	}
}
