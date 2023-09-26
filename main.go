package main

import "believer/movies/app"

func main() {
	err := app.SetupAndRunApp()

	if err != nil {
		panic(err)
	}
}
