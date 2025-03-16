package imgpack

import (
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
	"strings"

	_ "golang.org/x/image/webp"

	"archive/zip"
	"image"
	"io"
	"os"
	"path"
	"path/filepath"
	"slices"

	"github.com/VoileLab/goimgpack/internal/util"
)

// Img stores all the information of an image
type Img struct {
	// filename is the base name of the image file without the extension
	filename string

	// img is the image.Image object of the image
	img image.Image

	// imgType is the type of the image
	imgType string
}

func newImgByFilepath(filepath string) (*Img, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, util.Errorf("%w", err)
	}
	defer f.Close()

	return newImg(f, path.Base(filepath))
}

func newImg(r io.Reader, filename string) (*Img, error) {
	img, imgType, err := image.Decode(r)
	if err != nil {
		return nil, util.Errorf("%w", err)
	}

	// remove ext of filename
	filename = strings.TrimSuffix(filename, filepath.Ext(filename))

	return &Img{
		filename: filename,
		img:      img,
		imgType:  imgType,
	}, nil
}

func (img *Img) Clone() *Img {
	return &Img{
		filename: img.filename,
		img:      util.CloneImage(img.img),
		imgType:  img.imgType,
	}
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

	img, err := newImgByFilepath(filename)
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
			return nil, util.Errorf("%w", err)
		}
		defer rc.Close()

		img, err := newImg(rc, f.Name)
		if err != nil {
			return nil, util.Errorf("%w", err)
		}

		imgs = append(imgs, img)
	}

	return imgs, nil
}

func saveImgsAsZip(imgs []*Img, filepath string, prependDigit bool) error {
	zipFile, err := os.Create(filepath)
	if err != nil {
		return util.Errorf("%w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	imgLenDigits := util.CountDigits(len(imgs))
	for i, img := range imgs {
		filename := img.filename + ".jpg"
		if prependDigit {
			filename = util.PaddingZero(i, imgLenDigits) + "_" + filename
		}
		imgFile, err := zipWriter.Create(filename)
		if err != nil {
			return util.Errorf("%w", err)
		}

		err = jpeg.Encode(imgFile, img.img, &jpeg.Options{Quality: 90})
		if err != nil {
			return util.Errorf("%w", err)
		}
	}

	return nil
}
