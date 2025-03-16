package main

import (
	"os"

	"github.com/VoileLab/goimgpack/imgpack"
)

func main() {
	os.Setenv("FYNE_SCALE", "2")

	app := imgpack.NewImgpackApp()
	app.Run()
}
