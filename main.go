package main

import (
	"fmt"
	"log"
	"os"

	"github.com/VoileLab/goimgpack/imgpack"
)

func main() {
	scale := imgpack.GetPreferenceScale()
	log.Println("Read scale from preference:", scale)

	os.Setenv("FYNE_SCALE", fmt.Sprintf("%f", scale))

	app := imgpack.NewImgpackApp()
	app.Run()
}
