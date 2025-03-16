package util

import (
	"image"
	"image/draw"
)

// CloneImage clones an image.Image object
func CloneImage(src image.Image) image.Image {
	// Create a new RGBA image with the same bounds
	bounds := src.Bounds()
	clone := image.NewRGBA(bounds)

	// Copy the source image to the new one
	draw.Draw(clone, bounds, src, bounds.Min, draw.Src)

	return clone
}
