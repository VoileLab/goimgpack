package assets

import (
	_ "embed"
	_ "image/png"

	"bytes"
	"image"
)

//go:embed img_placeholder.png
var imgPlaceholderBs []byte

var ImgPlaceholder image.Image

func init() {
	img, _, err := image.Decode(bytes.NewReader(imgPlaceholderBs))
	if err != nil {
		panic(err)
	}
	ImgPlaceholder = img
}
