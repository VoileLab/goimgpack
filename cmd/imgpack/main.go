package main

import "github.com/VoileLab/goimgpack/imgpack"

func main() {
	err := imgpack.Main()
	if err != nil {
		panic(err)
	}
}
