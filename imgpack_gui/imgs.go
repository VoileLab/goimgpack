package imgpackgui

import (
	"archive/zip"
	"image"
	"io"
	"os"
	"path/filepath"

	"github.com/VoileLab/goimgpack/internal/util"
)

// Img stores all the information of an image
type Img struct {
	// uri is a local file URI
	uri string

	// basename is the base name of the image file
	basename string

	// img is the image.Image object of the image
	img image.Image
}

func newImg(uri string) (*Img, error) {
	f, err := os.Open(uri)
	if err != nil {
		return nil, util.Errorf("%w", err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, util.Errorf("%w", err)
	}

	return &Img{
		uri:      uri,
		basename: filepath.Base(uri),
		img:      img,
	}, nil
}

func saveImgsAsZip(imgs []*Img, uri string) error {
	zipFile, err := os.Create(uri)
	if err != nil {
		return util.Errorf("%w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	imgCnt := len(imgs)
	for i, img := range imgs {
		prefix := util.PaddingZero(i, imgCnt) + "_"
		imgFile, err := zipWriter.Create(prefix + img.basename)
		if err != nil {
			return util.Errorf("%w", err)
		}

		f, err := os.Open(img.uri)
		if err != nil {
			return util.Errorf("%w", err)
		}
		defer f.Close()

		_, err = io.Copy(imgFile, f)
		if err != nil {
			return util.Errorf("%w", err)
		}
	}

	return nil
}
