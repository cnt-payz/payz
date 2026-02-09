package main

import "github.com/cnt-payz/payz/crypto-service/internal/app"

func main() {
	app := app.Init()
	app.Run()
}
