package main

import "github.com/cnt-payz/payz/shop-gateway/internal/app"

func main() {
	app := app.Init()
	app.Run()
}
