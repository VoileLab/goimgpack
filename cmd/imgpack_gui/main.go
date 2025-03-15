package main

import imgpackgui "github.com/VoileLab/goimgpack/imgpack_gui"

func main() {
	err := imgpackgui.Main()
	if err != nil {
		panic(err)
	}
}
