package imgpackgui

import (
	_ "image/jpeg"
	_ "image/png"
	"log"

	_ "golang.org/x/image/webp"

	"archive/zip"
	"image"
	"io"
	"os"
	"path/filepath"
	"slices"

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

	// imgType is the type of the image
	imgType string
}

func newImg(uri string) (*Img, error) {
	f, err := os.Open(uri)
	if err != nil {
		return nil, util.Errorf("%w", err)
	}
	defer f.Close()

	img, imgType, err := image.Decode(f)
	if err != nil {
		return nil, util.Errorf("%w", err)
	}

	return &Img{
		uri:      uri,
		basename: filepath.Base(uri),
		img:      img,
		imgType:  imgType,
	}, nil
}

func readImgs(filename string) ([]*Img, error) {
	fileExt := filepath.Ext(filename)
	if slices.Contains(supportedArchiveExts, fileExt) {
		imgs, err := readImgsInZip(filename)
		if err != nil {
			return nil, util.Errorf("%w", err)
		}
		return imgs, nil
	}

	img, err := newImg(filename)
	if err != nil {
		return nil, util.Errorf("%w", err)
	}

	return []*Img{img}, nil
}

func readImgsInZip(filename string) ([]*Img, error) {
	r, err := zip.OpenReader(filename)
	if err != nil {
		return nil, util.Errorf("%w", err)
	}
	defer r.Close()

	imgs := make([]*Img, 0, len(r.File))
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}

		if !slices.Contains(supportedImageExts, filepath.Ext(f.Name)) {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			log.Println(err)
			continue
		}
		defer rc.Close()

		img, imgType, err := image.Decode(rc)
		if err != nil {
			log.Println(err)
			continue
		}

		imgs = append(imgs, &Img{
			uri:      f.Name,
			basename: filepath.Base(f.Name),
			img:      img,
			imgType:  imgType,
		})
	}

	return imgs, nil
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
