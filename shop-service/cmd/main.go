package main

import "github.com/cnt-payz/payz/shop-service/internal/app"

func main() {
	app := app.Init()
	app.Run()
}
