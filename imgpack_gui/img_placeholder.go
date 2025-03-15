package imgpackgui

import (
	"bytes"
	_ "embed"
	"image"
)

//go:embed img_placeholder.png
var imgPlaceholderBs []byte

var imgPlaceholder image.Image

func init() {
	img, _, err := image.Decode(bytes.NewReader(imgPlaceholderBs))
	if err != nil {
		panic(err)
	}
	imgPlaceholder = img
}
