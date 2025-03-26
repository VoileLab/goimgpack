package imgutil

import (
	"archive/zip"
	"bytes"
	"fmt"
	"image"
	"image/draw"
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

func NewImgByFilepath(filepath string) (*Image, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, util.Errorf("%w", err)
	}
	defer f.Close()

	return NewImg(f, path.Base(filepath))
}

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

func ReadImgs(filename string) ([]*Image, error) {
	fileExt := filepath.Ext(filename)
	if slices.Contains(SupportedArchiveExts, fileExt) {
		imgs, err := ReadImgsInZip(filename)
		if err != nil {
			return nil, util.Errorf("%w", err)
		}
		return imgs, nil
	}

	if fileExt == ".pdf" {
		imgs, err := ReadImgsInPDF(filename)
		if err != nil {
			return nil, util.Errorf("%w", err)
		}
		return imgs, nil
	}

	fileStat, err := os.Stat(filename)
	if err != nil {
		return nil, util.Errorf("%w", err)
	}

	if fileStat.IsDir() {
		imgs, err := ReadImgsInDir(filename)
		if err != nil {
			return nil, util.Errorf("%w", err)
		}
		return imgs, nil
	}

	img, err := NewImgByFilepath(filename)
	if err != nil {
		return nil, util.Errorf("%w", err)
	}

	return []*Image{img}, nil
}

func ReadImgsInDir(dirpath string) ([]*Image, error) {
	dir, err := os.ReadDir(dirpath)
	if err != nil {
		return nil, util.Errorf("%w", err)
	}

	imgs := make([]*Image, 0, len(dir))
	for _, entry := range dir {
		if entry.IsDir() {
			continue
		}

		img, err := NewImgByFilepath(filepath.Join(dirpath, entry.Name()))
		if err != nil {
			return nil, util.Errorf("%w", err)
		}

		imgs = append(imgs, img)
	}

	return imgs, nil
}

func ReadImgsInZip(filename string) ([]*Image, error) {
	r, err := zip.OpenReader(filename)
	if err != nil {
		return nil, util.Errorf("%w", err)
	}
	defer r.Close()

	imgs := make([]*Image, 0, len(r.File))
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}

		if !slices.Contains(SupportedImageExts, filepath.Ext(f.Name)) {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return nil, util.Errorf("%w", err)
		}
		defer rc.Close()

		// prevent directory in filename
		filename := strings.ReplaceAll(f.Name, "/", "_")

		img, err := NewImg(rc, filename)
		if err != nil {
			return nil, util.Errorf("%w", err)
		}

		imgs = append(imgs, img)
	}

	return imgs, nil
}

func ReadImgsInPDF(filename string) ([]*Image, error) {
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

	imgsMap := make(map[string]*Image)
	for idx, imgMap := range imgsInPDF {
		for jdx, imgReader := range imgMap {
			filename := fmt.Sprintf("%s_%d", util.PaddingZero(jdx, jdxMaxDigits), idx)

			img, err := NewImg(imgReader, filename)
			if err != nil {
				return nil, util.Errorf("%w", err)
			}

			imgsMap[filename] = img
		}
	}

	imgsKeys := slices.Collect(maps.Keys(imgsMap))
	slices.Sort(imgsKeys)

	imgs := make([]*Image, len(imgsKeys))
	for i, key := range imgsKeys {
		imgs[i] = imgsMap[key]
	}

	return imgs, nil
}

func SaveImg(img *Image, f io.Writer, quality int) error {
	err := jpeg.Encode(f, img.Img, &jpeg.Options{Quality: quality})
	if err != nil {
		return util.Errorf("%w", err)
	}

	return nil
}

func SaveImgsAsZip(imgs []*Image, f io.Writer, prependDigit bool, quality int) error {
	zipWriter := zip.NewWriter(f)
	defer zipWriter.Close()

	imgLenDigits := util.CountDigits(len(imgs))
	for i, img := range imgs {
		filename := img.Filename + ".jpg"
		if prependDigit {
			filename = util.PaddingZero(i, imgLenDigits) + "_" + filename
		}
		imgFile, err := zipWriter.Create(filename)
		if err != nil {
			return util.Errorf("%w", err)
		}

		err = jpeg.Encode(imgFile, img.Img, &jpeg.Options{Quality: quality})
		if err != nil {
			return util.Errorf("%w", err)
		}
	}

	return nil
}

func SaveImgsAsPDF(imgs []*Image, f io.Writer, quality int) error {
	imgsReader := make([]io.Reader, len(imgs))
	for i, img := range imgs {
		buf := new(bytes.Buffer)
		err := jpeg.Encode(buf, img.Img, &jpeg.Options{Quality: quality})
		if err != nil {
			return util.Errorf("%w", err)
		}

		imgsReader[i] = buf
	}

	err := api.ImportImages(nil, f, imgsReader, nil, nil)
	if err != nil {
		return util.Errorf("%w", err)
	}

	return nil
}
