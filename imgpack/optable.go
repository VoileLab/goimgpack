package imgpack

import (
	"image"
	"slices"

	"github.com/VoileLab/goimgpack/internal/imgutil"
	"github.com/disintegration/imaging"
)

type ImgsTable struct {
	selIdx *int
	imgs   []*imgutil.Image
}

func (t *ImgsTable) Select(idx int) {
	t.selIdx = &idx
}

func (t *ImgsTable) IsSelected() bool {
	return t.selIdx != nil
}

func (t *ImgsTable) GetSelected() *imgutil.Image {
	if t.selIdx == nil {
		return nil
	}

	return t.imgs[*t.selIdx]
}

func (t *ImgsTable) Unselect() {
	t.selIdx = nil
}

func (t *ImgsTable) Insert(imgs ...*imgutil.Image) {
	if t.selIdx == nil {
		t.imgs = append(t.imgs, imgs...)
		return
	}

	idx := *t.selIdx
	t.imgs = slices.Insert(t.imgs, idx+1, imgs...)
}

func (t *ImgsTable) Clear() {
	t.imgs = nil
}

func (t *ImgsTable) Remove(idx int) {
}

func (t *ImgsTable) Delete() {
	if t.selIdx == nil {
		return
	}

	idx := *t.selIdx
	t.imgs = slices.Delete(t.imgs, idx, idx+1)
}

func (t *ImgsTable) Duplicate() {
	if t.selIdx == nil {
		return
	}

	idx := *t.selIdx
	img := t.imgs[idx]

	newImg := img.Clone()

	t.imgs = slices.Insert(t.imgs, idx+1, newImg)
}

func (t *ImgsTable) MoveUp() {
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

func (t *ImgsTable) MoveDown() {
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

func (t *ImgsTable) Rotate() {
	if t.selIdx == nil {
		return
	}

	img := t.imgs[*t.selIdx]
	img.Img = imaging.Rotate90(img.Img)
}

func (t *ImgsTable) Cut() {
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
