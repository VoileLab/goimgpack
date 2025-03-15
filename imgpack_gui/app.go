package imgpackgui

import (
	"fmt"
	"goimgpack/internal/util"
	"image"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const appID = "com.mukyu.voile.imgpack"
const appTitle = "Image Packer"
const appVersion = "v0.1"

var appSize = fyne.NewSize(800, 600)

var supportedImageExts = []string{".png", ".jpg", ".jpeg"}

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

type ImgpackApp struct {
	fApp   fyne.App
	window fyne.Window

	stateBar      *widget.Label
	imgListWidget *widget.List
	imgShow       *canvas.Image

	selectedImgIdx *int

	imgs []*Img
}

func NewImgpackApp() *ImgpackApp {
	fApp := app.NewWithID(appID)
	window := fApp.NewWindow(appTitle)
	window.Resize(appSize)
	window.CenterOnScreen()

	retApp := &ImgpackApp{
		fApp:   fApp,
		window: window,
	}

	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.ContentAddIcon(), retApp.toolbarAddAction),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.DeleteIcon(), retApp.toolbarDeleteAction),
		widget.NewToolbarAction(theme.ContentCopyIcon(), retApp.toolbarDupAction),
		widget.NewToolbarAction(theme.MoveUpIcon(), retApp.toolbarMoveUpAction),
		widget.NewToolbarAction(theme.MoveDownIcon(), retApp.toolbarMoveDownAction),
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.HelpIcon(), func() {
			infoStr := fmt.Sprintf("%s %s", appTitle, appVersion)
			dialog := dialog.NewInformation("About", infoStr, window)
			dialog.Show()
		}),
	)

	imgListWidget := widget.NewList(
		func() int {
			return len(retApp.imgs)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Item")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(retApp.imgs[i].basename)
		},
	)
	imgListWidget.OnSelected = retApp.onSelectImageURI
	retApp.imgListWidget = imgListWidget

	imgShow := canvas.NewImageFromImage(imgPlaceholder)
	imgShow.FillMode = canvas.ImageFillContain
	retApp.imgShow = imgShow

	hSplit := container.NewHSplit(imgListWidget, imgShow)
	hSplit.SetOffset(0.25)

	stateBar := widget.NewLabel("Ready")
	retApp.stateBar = stateBar

	content := container.NewBorder(toolbar,
		stateBar, nil, nil, hSplit)

	window.SetContent(content)

	return retApp
}

func (app *ImgpackApp) Run() {
	app.window.ShowAndRun()
}

func (app *ImgpackApp) toolbarAddAction() {
	dlg := dialog.NewFileOpen(func(f fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, app.window)
			return
		}

		if f == nil {
			log.Println("Cancelled")
			return
		}

		if f.URI() == nil {
			log.Println("URI is nil")
			return
		}

		uri := f.URI().String()
		uri = strings.TrimPrefix(uri, "file://")

		img, err := newImg(uri)
		if err != nil {
			dialog.ShowError(err, app.window)
			return
		}

		if app.selectedImgIdx == nil {
			app.imgs = append(app.imgs, img)
		} else {
			idx := *app.selectedImgIdx
			app.imgs = slices.Insert(app.imgs, idx+1, img)
		}

		app.imgListWidget.Refresh()
	}, app.window)
	dlg.SetFilter(storage.NewExtensionFileFilter(supportedImageExts))
	dlg.Show()
}

func (app *ImgpackApp) onSelectImageURI(id widget.ListItemID) {
	app.selectedImgIdx = &id
	img := app.imgs[id]
	app.stateBar.SetText(img.uri)

	app.imgShow.Resource = nil
	app.imgShow.Image = img.img
	app.imgShow.Refresh()
}

func (app *ImgpackApp) toolbarDeleteAction() {
	if app.selectedImgIdx == nil {
		return
	}

	idx := *app.selectedImgIdx
	app.imgs = slices.Delete(app.imgs, idx, idx+1)
	app.imgListWidget.Refresh()

	if idx >= len(app.imgs) {
		app.selectedImgIdx = nil
		app.imgShow.Resource = nil
		app.imgShow.Image = imgPlaceholder
		app.imgShow.Refresh()
	} else {
		app.onSelectImageURI(idx)
	}
}

func (app *ImgpackApp) toolbarDupAction() {
	if app.selectedImgIdx == nil {
		return
	}

	idx := *app.selectedImgIdx
	img := app.imgs[idx]

	// Here we mean to copy the img, maybe we should write a function to copy img struct
	newImg, err := newImg(img.uri)
	if err != nil {
		dialog.ShowError(err, app.window)
		return
	}

	app.imgs = slices.Insert(app.imgs, idx+1, newImg)
	app.imgListWidget.Refresh()
}

func (app *ImgpackApp) toolbarMoveUpAction() {
	if app.selectedImgIdx == nil {
		return
	}

	idx := *app.selectedImgIdx
	if idx == 0 {
		return
	}

	app.imgs[idx], app.imgs[idx-1] = app.imgs[idx-1], app.imgs[idx]
	app.onSelectImageURI(idx)
	app.imgListWidget.Refresh()
}

func (app *ImgpackApp) toolbarMoveDownAction() {
	if app.selectedImgIdx == nil {
		return
	}

	idx := *app.selectedImgIdx
	if idx == len(app.imgs)-1 {
		return
	}

	app.imgs[idx], app.imgs[idx+1] = app.imgs[idx+1], app.imgs[idx]
	app.onSelectImageURI(idx)
	app.imgListWidget.Refresh()
}
