package imgutil

import (
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"

	"bytes"
	"image"
	"io"

	"github.com/VoileLab/goimgpack/internal/util"
	"github.com/disintegration/imaging"
)

var SupportedImageExts = []string{".png", ".jpg", ".jpeg", ".webp", ".bmp", ".tiff", ".gif"}

const formatJPEG = "jpeg"

func decodeImage(r io.Reader) (image.Image, string, error) {
	bs, err := io.ReadAll(r)
	if err != nil {
		return nil, "", util.Errorf("%w", err)
	}

	_, imgType, err := image.DecodeConfig(bytes.NewReader(bs))
	if err != nil {
		return nil, "", util.Errorf("%w", err)
	}

	var img image.Image

	if imgType == formatJPEG {
		// We should handle the orientation of the image
		img, err = imaging.Decode(bytes.NewReader(bs),
			imaging.AutoOrientation(true))
	} else {
		img, _, err = image.Decode(bytes.NewReader(bs))
	}

	if err != nil {
		return nil, "", util.Errorf("%w", err)
	}

	return img, imgType, nil
}
