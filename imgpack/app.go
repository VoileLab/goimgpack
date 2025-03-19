package imgpack

import (
	"fmt"
	"log"
	"net/url"
	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	dialogx "fyne.io/x/fyne/dialog"

	"github.com/disintegration/imaging"

	"github.com/VoileLab/goimgpack/imgpack/assets"
)

const appDescription = "A tool to pack images into an archive file."
const appURL = "https://github.com/VoileLab/goimgpack"

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
	fApp := app.New()
	meta := fApp.Metadata()

	mainWindow := fApp.NewWindow(meta.Name)
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
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.MediaReplayIcon(), retApp.toolbarRotateAction),
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.SettingsIcon(), func() {
			retApp.preferenceWindow.Show()
		}),
		widget.NewToolbarAction(theme.HelpIcon(), func() {
			docURL, _ := url.Parse(appURL)
			links := []*widget.Hyperlink{
				widget.NewHyperlink("Github", docURL),
			}

			dialogx.ShowAbout(appDescription, links, fApp, mainWindow)
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

	imgShow := canvas.NewImageFromImage(assets.ImgPlaceholder)
	imgShow.FillMode = canvas.ImageFillContain
	retApp.imgShow = imgShow

	hSplit := container.NewHSplit(imgListWidget, imgShow)
	hSplit.SetOffset(0.25)

	stateBar := widget.NewLabel("Ready")
	retApp.stateBar = stateBar

	content := container.NewBorder(toolbar,
		stateBar, nil, nil, hSplit)

	mainWindow.SetContent(content)

	addDigitCheck := widget.NewCheck("", setPreferencePrependDigit)
	addDigitCheck.SetChecked(getPreferencePrependDigit())

	jpgQualitySliderLabel := widget.NewLabel(
		fmt.Sprintf("JPG Quality: %d", getPreferenceJPGQuality()))

	jpgQualitySlider := widget.NewSlider(0, 100)
	jpgQualitySlider.Step = 1
	jpgQualitySlider.Value = float64(getPreferenceJPGQuality())
	jpgQualitySlider.OnChanged = func(v float64) {
		setPreferenceJPGQuality(int(v))
		jpgQualitySliderLabel.SetText(fmt.Sprintf("JPG Quality: %d", int(v)))
	}

	prefBody := container.New(layout.NewFormLayout(),
		widget.NewLabel("Add digit to filename"),
		addDigitCheck,
		jpgQualitySliderLabel,
		jpgQualitySlider,
	)

	preferenceWindow.SetContent(container.NewBorder(nil,
		widget.NewButton("Close", func() {
			retApp.preferenceWindow.Hide()
		}), nil, nil, prefBody))

	return retApp
}

func (iApp *ImgpackApp) Run() {
	iApp.mainWindow.ShowAndRun()
}

func (iApp *ImgpackApp) dropFiles(files []fyne.URI) {
	for _, file := range files {
		imgs, err := readImgs(file.Path())
		if err != nil {
			dialog.ShowError(err, iApp.mainWindow)
			continue
		}

		if iApp.selectedImgIdx == nil {
			iApp.imgs = append(iApp.imgs, imgs...)
		} else {
			idx := *iApp.selectedImgIdx
			iApp.imgs = slices.Insert(iApp.imgs, idx+1, imgs...)
		}
	}

	iApp.imgListWidget.Refresh()
}

func (iApp *ImgpackApp) onTabKey(e *fyne.KeyEvent) {
	switch e.Name {
	case fyne.KeyUp:
		if iApp.selectedImgIdx != nil {
			idx := *iApp.selectedImgIdx
			if idx > 0 {
				iApp.imgListWidget.Select(idx - 1)
			} else {
				iApp.imgListWidget.Select(len(iApp.imgs) - 1)
			}
		}
	case fyne.KeyDown:
		if iApp.selectedImgIdx != nil {
			idx := *iApp.selectedImgIdx
			if idx < len(iApp.imgs)-1 {
				iApp.imgListWidget.Select(idx + 1)
			} else {
				iApp.imgListWidget.Select(0)
			}
		}
	}
}

func (iApp *ImgpackApp) toolbarClearAction() {
	if len(iApp.imgs) == 0 {
		return
	}

	dialog.ShowConfirm("Clear all images", "Are you sure to clear all images?",
		func(b bool) {
			if b {
				iApp.imgs = []*Img{}
				iApp.selectedImgIdx = nil
			}
		},
		iApp.mainWindow)
}

func (iApp *ImgpackApp) toolbarAddAction() {
	dlg := dialog.NewFileOpen(func(f fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, iApp.mainWindow)
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
			dialog.ShowError(err, iApp.mainWindow)
			return
		}

		if iApp.selectedImgIdx == nil {
			iApp.imgs = append(iApp.imgs, imgs...)
		} else {
			idx := *iApp.selectedImgIdx
			iApp.imgs = slices.Insert(iApp.imgs, idx+1, imgs...)
		}

		iApp.imgListWidget.Refresh()
	}, iApp.mainWindow)
	dlg.SetFilter(storage.NewExtensionFileFilter(supportedAddExts))
	dlg.Resize(fyne.NewSize(600, 600))
	dlg.Show()
}

func (iApp *ImgpackApp) toolbarDownloadAction() {
	if iApp.selectedImgIdx == nil {
		return
	}

	img := iApp.imgs[*iApp.selectedImgIdx]
	dlg := dialog.NewFileSave(func(f fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, iApp.mainWindow)
			return
		}

		if f == nil {
			log.Println("Cancelled")
			return
		}

		err = saveImg(img, f)
		if err != nil {
			dialog.ShowError(err, iApp.mainWindow)
			return
		}
		f.Close()

		iApp.stateBar.SetText("Saved successfully")
	}, iApp.mainWindow)

	dlg.SetFileName(img.filename + ".jpg")
	dlg.SetFilter(storage.NewExtensionFileFilter(supportedImageExts))
	dlg.Resize(fyne.NewSize(600, 600))
	dlg.Show()
}

func (iApp *ImgpackApp) onSelectImageURI(id widget.ListItemID) {
	iApp.selectedImgIdx = &id
	img := iApp.imgs[id]

	stateText := fmt.Sprintf("Selected: %s - type: %s", img.filename, img.imgType)
	iApp.stateBar.SetText(stateText)

	iApp.imgShow.Resource = nil
	iApp.imgShow.Image = img.img
	iApp.imgShow.Refresh()
}

func (iApp *ImgpackApp) toolbarDeleteAction() {
	if iApp.selectedImgIdx == nil {
		return
	}

	idx := *iApp.selectedImgIdx
	iApp.imgs = slices.Delete(iApp.imgs, idx, idx+1)
	iApp.imgListWidget.Refresh()

	if idx >= len(iApp.imgs) {
		iApp.selectedImgIdx = nil
		iApp.imgShow.Resource = nil
		iApp.imgShow.Image = assets.ImgPlaceholder
		iApp.imgShow.Refresh()
	} else {
		iApp.onSelectImageURI(idx)
	}
}

func (iApp *ImgpackApp) toolbarDupAction() {
	if iApp.selectedImgIdx == nil {
		return
	}

	idx := *iApp.selectedImgIdx
	img := iApp.imgs[idx]

	newImg := img.Clone()

	iApp.imgs = slices.Insert(iApp.imgs, idx+1, newImg)
	iApp.imgListWidget.Refresh()
}

func (iApp *ImgpackApp) toolbarMoveUpAction() {
	if iApp.selectedImgIdx == nil {
		return
	}

	idx := *iApp.selectedImgIdx
	if idx == 0 {
		iApp.imgs = append(iApp.imgs[1:], iApp.imgs[0])
		iApp.imgListWidget.Select(len(iApp.imgs) - 1)
		return
	}

	iApp.imgs[idx], iApp.imgs[idx-1] = iApp.imgs[idx-1], iApp.imgs[idx]
	iApp.imgListWidget.Select(idx - 1)
}

func (iApp *ImgpackApp) toolbarMoveDownAction() {
	if iApp.selectedImgIdx == nil {
		return
	}

	idx := *iApp.selectedImgIdx
	if idx == len(iApp.imgs)-1 {
		iApp.imgs = append([]*Img{iApp.imgs[len(iApp.imgs)-1]}, iApp.imgs[:len(iApp.imgs)-1]...)
		iApp.imgListWidget.Select(0)
		return
	}

	iApp.imgs[idx], iApp.imgs[idx+1] = iApp.imgs[idx+1], iApp.imgs[idx]
	iApp.imgListWidget.Select(idx + 1)
}

func (iApp *ImgpackApp) toolbarSaveAction() {
	if len(iApp.imgs) == 0 {
		iApp.stateBar.SetText("No image to save")
		return
	}

	dlg := dialog.NewFileSave(func(f fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, iApp.mainWindow)
			return
		}

		if f == nil {
			log.Println("Cancelled")
			return
		}

		err = saveImgsAsZip(iApp.imgs, f,
			getPreferencePrependDigit(),
			getPreferenceJPGQuality())
		if err != nil {
			dialog.ShowError(err, iApp.mainWindow)
			return
		}
		f.Close()

		iApp.stateBar.SetText("Saved successfully")
	}, iApp.mainWindow)

	dlg.SetFileName("output.cbz")
	dlg.SetFilter(storage.NewExtensionFileFilter(supportedArchiveExts))
	dlg.Resize(fyne.NewSize(600, 600))
	dlg.Show()
}

func (iApp *ImgpackApp) toolbarRotateAction() {
	if iApp.selectedImgIdx == nil {
		return
	}

	img := iApp.imgs[*iApp.selectedImgIdx]
	img.img = imaging.Rotate90(img.img)
	iApp.imgShow.Image = img.img
	iApp.imgShow.Refresh()
}
