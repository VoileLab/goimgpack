package assets

import (
	_ "embed"
	"image/color"
	"image/png"

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

var DisabledAsPdfIcon fyne.Resource

//go:embed description.md
var AppDescription string

func init() {
	img, _, err := image.Decode(bytes.NewReader(imgPlaceholderBs))
	if err != nil {
		panic(err)
	}
	ImgPlaceholder = img

	AsPdfIcon = fyne.NewStaticResource(
		"picture_as_pdf.png", pictureAsPdfBs)

	img, _, err = image.Decode(bytes.NewReader(pictureAsPdfBs))
	if err != nil {
		panic(err)
	}

	grayImg := toGrayImage(img)

	var buf bytes.Buffer
	if err := png.Encode(&buf, grayImg); err != nil {
		panic(err)
	}

	DisabledAsPdfIcon = fyne.NewStaticResource(
		"picture_as_pdf_disabled.png", buf.Bytes())
}

func toGrayImage(img image.Image) *image.RGBA {
	bounds := img.Bounds()
	grayImg := image.NewRGBA(img.Bounds())
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			_, _, _, a := img.At(x, y).RGBA()
			r, g, b, _ := color.GrayModel.Convert(img.At(x, y)).RGBA()
			grayImg.Set(x, y, color.RGBA{
				R: uint8(r),
				G: uint8(g),
				B: uint8(b),
				A: uint8(a) >> 2,
			})
		}
	}

	return grayImg
}
