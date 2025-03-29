package imgutil

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/VoileLab/goimgpack/internal/util"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// ReadImgsInFile reads images from a file
func ReadImgsInFile(f io.Reader, filename string) ([]*Image, error) {
	fileExt := filepath.Ext(filename)
	if slices.Contains(SupportedArchiveExts, fileExt) {
		imgs, err := ReadImgsInZip(f)
		if err != nil {
			return nil, util.Errorf("%w", err)
		}
		return imgs, nil
	}

	if slices.Contains(SupportedPDFExts, fileExt) {
		imgs, err := ReadImgsInPDF(f)
		if err != nil {
			return nil, util.Errorf("%w", err)
		}
		return imgs, nil
	}

	img, err := NewImg(f, filename)
	if err != nil {
		return nil, util.Errorf("%w", err)
	}

	return []*Image{img}, nil
}

// ReadImgsInDir reads images in a directory not recursively
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

// ReadImgsInZip reads images in a zip file
func ReadImgsInZip(f io.Reader) ([]*Image, error) {
	bs, err := io.ReadAll(f)
	if err != nil {
		return nil, util.Errorf("%w", err)
	}

	r, err := zip.NewReader(bytes.NewReader(bs), int64(len(bs)))
	if err != nil {
		return nil, util.Errorf("%w", err)
	}

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

// ReadImgsInPDF reads images in a PDF file
func ReadImgsInPDF(f io.Reader) ([]*Image, error) {
	conf := model.NewDefaultConfiguration()
	conf.ValidationMode = model.ValidationRelaxed

	pdfBuf := new(bytes.Buffer)
	_, err := io.Copy(pdfBuf, f)
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
