package util

import (
	"image"
	"image/draw"
)

// CloneImage clones an image.Image object
func CloneImage(src image.Image) image.Image {
	bounds := src.Bounds()
	clone := image.NewRGBA(bounds)
	draw.Draw(clone, bounds, src, bounds.Min, draw.Src)

	return clone
}
