package imgpack

import (
	"image"
	"slices"

	"github.com/VoileLab/goimgpack/internal/imgutil"
	"github.com/disintegration/imaging"
)

type OPTable struct {
	selIdx *int
	imgs   []*imgutil.Image
}

func (t *OPTable) Select(idx int) {
	t.selIdx = &idx
}

func (t *OPTable) IsSelected() bool {
	return t.selIdx != nil
}

func (t *OPTable) GetSelected() *imgutil.Image {
	if t.selIdx == nil {
		return nil
	}

	return t.imgs[*t.selIdx]
}

func (t *OPTable) Unselect() {
	t.selIdx = nil
}

func (t *OPTable) Insert(imgs ...*imgutil.Image) {
	if t.selIdx == nil {
		t.imgs = append(t.imgs, imgs...)
		return
	}

	idx := *t.selIdx
	t.imgs = slices.Insert(t.imgs, idx+1, imgs...)
}

func (t *OPTable) Clear() {
	t.imgs = nil
}

func (t *OPTable) Remove(idx int) {
}

func (t *OPTable) Delete() {
	if t.selIdx == nil {
		return
	}

	idx := *t.selIdx
	t.imgs = slices.Delete(t.imgs, idx, idx+1)
}

func (t *OPTable) Duplicate() {
	if t.selIdx == nil {
		return
	}

	idx := *t.selIdx
	img := t.imgs[idx]

	newImg := img.Clone()

	t.imgs = slices.Insert(t.imgs, idx+1, newImg)
}

func (t *OPTable) MoveUp() {
	if t.selIdx == nil {
		return
	}

	idx := *t.selIdx
	if idx == 0 {
		t.imgs = append(t.imgs[:1], t.imgs[0:]...)
		idx = len(t.imgs) - 1
		t.selIdx = &idx
		return
	}

	t.imgs[idx], t.imgs[idx-1] = t.imgs[idx-1], t.imgs[idx]
	idx = idx - 1
	t.selIdx = &idx
}

func (t *OPTable) MoveDown() {
	if t.selIdx == nil {
		return
	}

	idx := *t.selIdx
	if idx == len(t.imgs)-1 {
		t.imgs = append([]*imgutil.Image{t.imgs[len(t.imgs)-1]}, t.imgs[:len(t.imgs)-1]...)
		idx = 0
		t.selIdx = &idx
		return
	}

	t.imgs[idx], t.imgs[idx+1] = t.imgs[idx+1], t.imgs[idx]
	idx = idx + 1
	t.selIdx = &idx
}

func (t *OPTable) Rotate() {
	if t.selIdx == nil {
		return
	}

	img := t.imgs[*t.selIdx]
	img.Img = imaging.Rotate90(img.Img)
}

func (t *OPTable) Cut() {
	if t.selIdx == nil {
		return
	}

	idx := *t.selIdx
	img := t.imgs[idx]
	filename := img.Filename
	imgType := img.Type

	imgWidth := img.Img.Bounds().Dx()
	imgHeight := img.Img.Bounds().Dy()

	spWidth := imgWidth / 2

	img1 := imaging.Crop(img.Img, image.Rect(0, 0, spWidth, imgHeight))
	img2 := imaging.Crop(img.Img, image.Rect(spWidth, 0, imgWidth, imgHeight))

	img.Filename = filename + "_1"
	img.Img = img1

	newImg := &imgutil.Image{
		Filename: filename + "_2",
		Img:      img2,
		Type:     imgType,
	}

	t.imgs = slices.Insert(t.imgs, idx+1, newImg)
}
