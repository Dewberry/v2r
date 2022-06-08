package main

import (
	"app/handlers"
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	fmt.Println("Welcome to the fcast-dashboard API")
	// Instantiate echo
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/ping", handlers.Ping())
	e.GET("/read-data", handlers.ReadData())

	e.Logger.Fatal(e.Start(":5000"))
}
