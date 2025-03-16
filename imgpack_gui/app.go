package imgpackgui

import (
	"fmt"
	"log"
	"slices"

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

var appSize = fyne.NewSize(1000, 800)

var supportedImageExts = []string{".png", ".jpg", ".jpeg", ".webp"}
var supportedArchiveExts = []string{".zip", ".cbz"}

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

	window.SetOnDropped(func(p fyne.Position, u []fyne.URI) {
		retApp.dropFiles(u)
	})

	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.DocumentSaveIcon(), retApp.toolbarSaveAction),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.ContentAddIcon(), retApp.toolbarAddAction),
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
			o.(*widget.Label).SetText(retApp.imgs[i].filename)
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

func (app *ImgpackApp) dropFiles(files []fyne.URI) {
	for _, file := range files {
		img, err := newImgByFilepath(file.Path())
		if err != nil {
			dialog.ShowError(err, app.window)
			continue
		}

		app.imgs = append(app.imgs, img)
	}

	app.imgListWidget.Refresh()
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

		filepath := f.URI().Path()
		imgs, err := readImgs(filepath)
		if err != nil {
			dialog.ShowError(err, app.window)
			return
		}

		if app.selectedImgIdx == nil {
			app.imgs = append(app.imgs, imgs...)
		} else {
			idx := *app.selectedImgIdx
			app.imgs = slices.Insert(app.imgs, idx+1, imgs...)
		}

		app.imgListWidget.Refresh()
	}, app.window)
	dlg.SetFilter(storage.NewExtensionFileFilter(append(supportedImageExts, supportedArchiveExts...)))
	dlg.Resize(fyne.NewSize(600, 600))
	dlg.Show()
}

func (app *ImgpackApp) onSelectImageURI(id widget.ListItemID) {
	app.selectedImgIdx = &id
	img := app.imgs[id]

	stateText := fmt.Sprintf("Selected: %s - type: %s", img.filename, img.imgType)
	app.stateBar.SetText(stateText)

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

	newImg := img.Clone()

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

func (app *ImgpackApp) toolbarSaveAction() {
	if len(app.imgs) == 0 {
		app.stateBar.SetText("No image to save")
		return
	}

	dlg := dialog.NewFileSave(func(f fyne.URIWriteCloser, err error) {
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

		filepath := f.URI().Path()
		err = saveImgsAsZip(app.imgs, filepath)
		if err != nil {
			dialog.ShowError(err, app.window)
			return
		}

		app.stateBar.SetText("Saved successfully")
	}, app.window)

	dlg.SetFileName("output.cbz")
	dlg.SetFilter(storage.NewExtensionFileFilter(supportedArchiveExts))
	dlg.Resize(fyne.NewSize(600, 600))
	dlg.Show()
}
