package assets

import (
	_ "embed"
	_ "image/png"

	"bytes"
	"image"

	"fyne.io/fyne/v2"
)

//go:embed img_placeholder.png
var imgPlaceholderBs []byte

var ImgPlaceholder image.Image

//go:embed picture_as_pdf.png
var pictureAsPdfBs []byte

var AsPdfIcon fyne.Resource

func init() {
	img, _, err := image.Decode(bytes.NewReader(imgPlaceholderBs))
	if err != nil {
		panic(err)
	}
	ImgPlaceholder = img

	AsPdfIcon = fyne.NewStaticResource(
		"picture_as_pdf.png", pictureAsPdfBs)
}
