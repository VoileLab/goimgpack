package imgutil

import (
	"archive/zip"
	"bytes"
	"image/jpeg"
	"io"

	"github.com/VoileLab/goimgpack/internal/util"
	"github.com/pdfcpu/pdfcpu/pkg/api"
)

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
