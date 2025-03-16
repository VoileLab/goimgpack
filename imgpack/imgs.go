package imgpack

import (
	_ "image/png"

	_ "golang.org/x/image/webp"

	"archive/zip"
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"maps"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/VoileLab/goimgpack/internal/util"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
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

	if fileExt == ".pdf" {
		imgs, err := readImgsInPDF(filename)
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

func saveImg(img *Img, filepath string) error {
	f, err := os.Create(filepath)
	if err != nil {
		return util.Errorf("%w", err)
	}
	defer f.Close()

	err = jpeg.Encode(f, img.img, &jpeg.Options{Quality: 90})
	if err != nil {
		return util.Errorf("%w", err)
	}

	return nil
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

func readImgsInPDF(filename string) ([]*Img, error) {
	conf := model.NewDefaultConfiguration()
	conf.ValidationMode = model.ValidationRelaxed

	pdfFile, err := os.Open(filename)
	if err != nil {
		return nil, util.Errorf("%w", err)
	}
	defer pdfFile.Close()

	pdfBuf := new(bytes.Buffer)
	_, err = io.Copy(pdfBuf, pdfFile)
	if err != nil {
		return nil, util.Errorf("%w", err)
	}

	pdfReader := bytes.NewReader(pdfBuf.Bytes())

	allPages, err := api.PageCount(pdfReader, conf)
	if err != nil {
		return nil, util.Errorf("%w", err)
	}

	allPagesStr := make([]string, allPages)
	for i := range allPages {
		allPagesStr[i] = fmt.Sprintf("%d", i+1)
	}

	imgsInPDF, err := api.ExtractImagesRaw(pdfReader, allPagesStr, conf)
	if err != nil {
		return nil, util.Errorf("%w", err)
	}

	jdxMax := 0
	for _, imgMap := range imgsInPDF {
		for jdx := range imgMap {
			jdxMax = max(jdxMax, jdx)
		}
	}
	jdxMaxDigits := util.CountDigits(jdxMax)

	imgsMap := make(map[string]*Img)
	for idx, imgMap := range imgsInPDF {
		for jdx, imgReader := range imgMap {
			filename := fmt.Sprintf("%s_%d", util.PaddingZero(jdx, jdxMaxDigits), idx)

			img, err := newImg(imgReader, filename)
			if err != nil {
				return nil, util.Errorf("%w", err)
			}

			imgsMap[filename] = img
		}
	}

	imgsKeys := slices.Collect(maps.Keys(imgsMap))
	slices.Sort(imgsKeys)

	imgs := make([]*Img, len(imgsKeys))
	for i, key := range imgsKeys {
		imgs[i] = imgsMap[key]
	}

	return imgs, nil
}
