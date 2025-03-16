package imgpack

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
var supportedPDFExts = []string{".pdf"}
var supportedAddExts = slices.Concat(supportedImageExts, supportedArchiveExts, supportedPDFExts)

type ImgpackApp struct {
	fApp             fyne.App
	mainWindow       fyne.Window
	preferenceWindow fyne.Window

	stateBar      *widget.Label
	imgListWidget *widget.List
	imgShow       *canvas.Image

	selectedImgIdx *int

	imgs []*Img
}

func NewImgpackApp() *ImgpackApp {
	fApp := app.NewWithID(appID)

	mainWindow := fApp.NewWindow(appTitle)
	mainWindow.Resize(appSize)
	mainWindow.CenterOnScreen()

	preferenceWindow := fApp.NewWindow("Preferences")
	preferenceWindow.Resize(fyne.NewSize(400, 300))
	preferenceWindow.SetFixedSize(true)
	preferenceWindow.CenterOnScreen()
	preferenceWindow.SetCloseIntercept(func() {
		preferenceWindow.Hide()
	})

	retApp := &ImgpackApp{
		fApp:             fApp,
		mainWindow:       mainWindow,
		preferenceWindow: preferenceWindow,
	}

	mainWindow.SetOnDropped(func(p fyne.Position, u []fyne.URI) {
		retApp.dropFiles(u)
	})

	mainWindow.Canvas().SetOnTypedKey(retApp.onTabKey)

	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.DocumentCreateIcon(), retApp.toolbarClearAction),
		widget.NewToolbarAction(theme.DocumentSaveIcon(), retApp.toolbarSaveAction),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.ContentAddIcon(), retApp.toolbarAddAction),
		widget.NewToolbarAction(theme.DeleteIcon(), retApp.toolbarDeleteAction),
		widget.NewToolbarAction(theme.ContentCopyIcon(), retApp.toolbarDupAction),
		widget.NewToolbarAction(theme.MoveUpIcon(), retApp.toolbarMoveUpAction),
		widget.NewToolbarAction(theme.MoveDownIcon(), retApp.toolbarMoveDownAction),
		widget.NewToolbarAction(theme.DownloadIcon(), retApp.toolbarDownloadAction),
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.SettingsIcon(), func() {
			retApp.preferenceWindow.Show()
		}),
		widget.NewToolbarAction(theme.HelpIcon(), func() {
			infoStr := fmt.Sprintf("%s %s", appTitle, appVersion)
			dialog := dialog.NewInformation("About", infoStr, mainWindow)
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

	mainWindow.SetContent(content)

	addDigitCheck := widget.NewCheck("Add digit to filename", func(b bool) {
		fApp.Preferences().SetBool("add_digit", b)
	})
	addDigitCheck.SetChecked(fApp.Preferences().Bool("add_digit"))

	prefBody := container.NewVBox(
		addDigitCheck,
	)

	preferenceWindow.SetContent(container.NewBorder(nil,
		widget.NewButton("Close", func() {
			retApp.preferenceWindow.Hide()
		}), nil, nil, prefBody))

	return retApp
}

func (app *ImgpackApp) Run() {
	app.mainWindow.ShowAndRun()
}

func (app *ImgpackApp) dropFiles(files []fyne.URI) {
	for _, file := range files {
		img, err := newImgByFilepath(file.Path())
		if err != nil {
			dialog.ShowError(err, app.mainWindow)
			continue
		}

		app.imgs = append(app.imgs, img)
	}

	app.imgListWidget.Refresh()
}

func (app *ImgpackApp) onTabKey(e *fyne.KeyEvent) {
	switch e.Name {
	case fyne.KeyUp:
		if app.selectedImgIdx != nil {
			idx := *app.selectedImgIdx
			if idx > 0 {
				app.imgListWidget.Select(idx - 1)
			} else {
				app.imgListWidget.Select(len(app.imgs) - 1)
			}
		}
	case fyne.KeyDown:
		if app.selectedImgIdx != nil {
			idx := *app.selectedImgIdx
			if idx < len(app.imgs)-1 {
				app.imgListWidget.Select(idx + 1)
			} else {
				app.imgListWidget.Select(0)
			}
		}
	}
}

func (app *ImgpackApp) toolbarClearAction() {
	if len(app.imgs) == 0 {
		return
	}

	dialog.ShowConfirm("Clear all images", "Are you sure to clear all images?",
		func(b bool) {
			if b {
				app.imgs = []*Img{}
			}
		},
		app.mainWindow)
}

func (app *ImgpackApp) toolbarAddAction() {
	dlg := dialog.NewFileOpen(func(f fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, app.mainWindow)
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
			dialog.ShowError(err, app.mainWindow)
			return
		}

		if app.selectedImgIdx == nil {
			app.imgs = append(app.imgs, imgs...)
		} else {
			idx := *app.selectedImgIdx
			app.imgs = slices.Insert(app.imgs, idx+1, imgs...)
		}

		app.imgListWidget.Refresh()
	}, app.mainWindow)
	dlg.SetFilter(storage.NewExtensionFileFilter(supportedAddExts))
	dlg.Resize(fyne.NewSize(600, 600))
	dlg.Show()
}

func (app *ImgpackApp) toolbarDownloadAction() {
	if app.selectedImgIdx == nil {
		return
	}

	img := app.imgs[*app.selectedImgIdx]
	dlg := dialog.NewFileSave(func(f fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, app.mainWindow)
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
		err = saveImg(img, filepath)
		if err != nil {
			dialog.ShowError(err, app.mainWindow)
			return
		}

		app.stateBar.SetText("Saved successfully")
	}, app.mainWindow)

	dlg.SetFileName(img.filename + ".jpg")
	dlg.SetFilter(storage.NewExtensionFileFilter(supportedImageExts))
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
		app.imgs = append(app.imgs[1:], app.imgs[0])
		app.imgListWidget.Select(len(app.imgs) - 1)
		return
	}

	app.imgs[idx], app.imgs[idx-1] = app.imgs[idx-1], app.imgs[idx]
	app.imgListWidget.Select(idx - 1)
}

func (app *ImgpackApp) toolbarMoveDownAction() {
	if app.selectedImgIdx == nil {
		return
	}

	idx := *app.selectedImgIdx
	if idx == len(app.imgs)-1 {
		app.imgs = append([]*Img{app.imgs[len(app.imgs)-1]}, app.imgs[:len(app.imgs)-1]...)
		app.imgListWidget.Select(0)
		return
	}

	app.imgs[idx], app.imgs[idx+1] = app.imgs[idx+1], app.imgs[idx]
	app.imgListWidget.Select(idx + 1)
}

func (app *ImgpackApp) toolbarSaveAction() {
	if len(app.imgs) == 0 {
		app.stateBar.SetText("No image to save")
		return
	}

	dlg := dialog.NewFileSave(func(f fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, app.mainWindow)
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
		err = saveImgsAsZip(app.imgs, filepath, app.fApp.Preferences().Bool("add_digit"))
		if err != nil {
			dialog.ShowError(err, app.mainWindow)
			return
		}

		app.stateBar.SetText("Saved successfully")
	}, app.mainWindow)

	dlg.SetFileName("output.cbz")
	dlg.SetFilter(storage.NewExtensionFileFilter(supportedArchiveExts))
	dlg.Resize(fyne.NewSize(600, 600))
	dlg.Show()
}
