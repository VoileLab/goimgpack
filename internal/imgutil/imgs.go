package imgutil

import (
	"image"
	"image/draw"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/VoileLab/goimgpack/internal/util"
)

var SupportedArchiveExts = []string{".zip", ".cbz"}
var SupportedPDFExts = []string{".pdf"}

// Image stores all the information of an image
type Image struct {
	// Filename is the base name of the image file without the extension
	Filename string

	// Img is the image.Image object of the image
	Img image.Image

	// Type is the type of the image
	Type string
}

// NewImgByFilepath creates an Image object from a file path
func NewImgByFilepath(filepath string) (*Image, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, util.Errorf("%w", err)
	}
	defer f.Close()

	return NewImg(f, path.Base(filepath))
}

// NewImg creates an Image object from an io.Reader
func NewImg(r io.Reader, filename string) (*Image, error) {
	img, imgType, err := decodeImage(r)
	if err != nil {
		return nil, util.Errorf("%w", err)
	}

	// remove ext of filename
	filename = strings.TrimSuffix(filename, filepath.Ext(filename))

	return &Image{
		Filename: filename,
		Img:      img,
		Type:     imgType,
	}, nil
}

// Clone deep copies the image
func (img *Image) Clone() *Image {
	bounds := img.Img.Bounds()
	clone := image.NewRGBA(bounds)
	draw.Draw(clone, bounds, img.Img, bounds.Min, draw.Src)

	return &Image{
		Filename: img.Filename,
		Img:      clone,
		Type:     img.Type,
	}
}
