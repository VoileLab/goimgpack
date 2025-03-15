package main

import imgpackgui "goimgpack/imgpack_gui"

func main() {
	err := imgpackgui.Main()
	if err != nil {
		panic(err)
	}
}
