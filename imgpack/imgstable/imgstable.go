package imgstable

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

func New() *ImgsTable {
	return &ImgsTable{}
}

func (t *ImgsTable) Len() int {
	return len(t.imgs)
}

func (t *ImgsTable) Get(idx int) *imgutil.Image {
	return t.imgs[idx]
}

func (t *ImgsTable) GetImgs() []*imgutil.Image {
	return t.imgs
}

func (t *ImgsTable) Select(idx int) {
	if idx < 0 || idx >= len(t.imgs) {
		return
	}

	t.selIdx = &idx
}

func (t *ImgsTable) IsSelected() bool {
	return t.selIdx != nil
}

func (t *ImgsTable) GetSelectedIdx() int {
	return *t.selIdx
}

func (t *ImgsTable) GetSelectedImg() *imgutil.Image {
	if t.selIdx == nil {
		return nil
	}

	return t.imgs[*t.selIdx]
}

func (t *ImgsTable) Unselect() {
	t.selIdx = nil
}

// Insert inserts images at the selected position of the table.
func (t *ImgsTable) Insert(imgs ...*imgutil.Image) {
	if t.selIdx == nil {
		t.imgs = append(t.imgs, imgs...)
		return
	}

	idx := *t.selIdx
	t.imgs = slices.Insert(t.imgs, idx+1, imgs...)
}

// Clear removes all images from the table.
func (t *ImgsTable) Clear() {
	t.imgs = nil
}

// Delete removes the selected image from the table.
func (t *ImgsTable) Delete() {
	if t.selIdx == nil {
		return
	}

	idx := *t.selIdx
	t.imgs = slices.Delete(t.imgs, idx, idx+1)

	if idx >= len(t.imgs) {
		t.selIdx = nil
	}
}

// Duplicate duplicates the selected image.
func (t *ImgsTable) Duplicate() {
	if t.selIdx == nil {
		return
	}

	idx := *t.selIdx
	img := t.imgs[idx]

	newImg := img.Clone()

	t.imgs = slices.Insert(t.imgs, idx+1, newImg)
}

// MoveUp moves the selected image up.
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

// MoveDown moves the selected image down.
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

// Rotate rotates the selected image 90 degrees clockwise.
func (t *ImgsTable) Rotate() {
	if t.selIdx == nil {
		return
	}

	img := t.imgs[*t.selIdx]
	img.Img = imaging.Rotate90(img.Img)
}

// Cut cuts the selected image in half and
// inserts the second half after the selected image.
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
