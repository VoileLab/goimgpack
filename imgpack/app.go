package imgpack

import (
	"fmt"
	"image"
	"log"
	"net/url"
	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	dialogx "fyne.io/x/fyne/dialog"

	"github.com/disintegration/imaging"

	"github.com/VoileLab/goimgpack/imgpack/assets"
	"github.com/VoileLab/goimgpack/internal/imgutil"
)

const appDescription = "A tool to pack images into an archive file."
const appURL = "https://github.com/VoileLab/goimgpack"

var appSize = fyne.NewSize(1000, 800)

type ImgpackApp struct {
	fApp       fyne.App
	mainWindow fyne.Window

	toolbar *widget.Toolbar

	stateBar      *widget.Label
	imgListWidget *widget.List
	imgShow       *canvas.Image

	selectedImgIdx *int

	imgs []*imgutil.Image

	enableOnSelectImageEnables []Enablable

	// reading progress dialog
	readingImagesDlg dialog.Dialog
	savingDlg        dialog.Dialog
}

func NewImgpackApp() *ImgpackApp {
	fApp := app.New()
	meta := fApp.Metadata()

	mainWindow := fApp.NewWindow(meta.Name)
	mainWindow.Resize(appSize)
	mainWindow.CenterOnScreen()

	retApp := &ImgpackApp{
		fApp:       fApp,
		mainWindow: mainWindow,
	}

	mainWindow.SetOnDropped(func(p fyne.Position, u []fyne.URI) {
		retApp.dropFiles(u)
	})

	mainWindow.Canvas().SetOnTypedKey(retApp.onTabKey)

	retApp.enableOnSelectImageEnables = []Enablable{}

	retApp.setupDialogs()
	retApp.setupMenu()
	retApp.setupToolbar()
	retApp.setupContent()

	for _, action := range retApp.enableOnSelectImageEnables {
		action.Disable()
	}

	return retApp
}

func (iApp *ImgpackApp) setupDialogs() {
	readingImagesDlg := dialog.NewCustomWithoutButtons(
		"Reading images...", widget.NewProgressBarInfinite(), iApp.mainWindow)
	iApp.readingImagesDlg = readingImagesDlg

	savingDlg := dialog.NewCustomWithoutButtons(
		"Saving...", widget.NewProgressBarInfinite(), iApp.mainWindow)
	iApp.savingDlg = savingDlg
}

func (iApp *ImgpackApp) setupMenu() {
	addImgsMenuItem := fyne.NewMenuItem("Add", iApp.addAction)
	addImgsMenuItem.Icon = theme.ContentAddIcon()

	delImgsMenuItem := fyne.NewMenuItem("Delete", iApp.deleteAction)
	dupImgsMenuItem := fyne.NewMenuItem("Duplicate", iApp.dupAction)
	moveUpImgsMenuItem := fyne.NewMenuItem("Move Up", iApp.moveUpAction)
	moveDownImgsMenuItem := fyne.NewMenuItem("Move Down", iApp.moveDownAction)
	downloadImgsMenuItem := fyne.NewMenuItem("Download", iApp.downloadAction)
	rotateImgsMenuItem := fyne.NewMenuItem("Rotate", iApp.rotateAction)
	cutImgMenuItem := fyne.NewMenuItem("Cut", iApp.cutAction)

	iApp.enableOnSelectImageEnables = append(
		iApp.enableOnSelectImageEnables,
		&EnablableWrapMenuItem{addImgsMenuItem},
		&EnablableWrapMenuItem{delImgsMenuItem},
		&EnablableWrapMenuItem{dupImgsMenuItem},
		&EnablableWrapMenuItem{moveUpImgsMenuItem},
		&EnablableWrapMenuItem{moveDownImgsMenuItem},
		&EnablableWrapMenuItem{downloadImgsMenuItem},
		&EnablableWrapMenuItem{rotateImgsMenuItem},
		&EnablableWrapMenuItem{cutImgMenuItem},
	)

	menu := fyne.NewMainMenu(
		fyne.NewMenu("File",
			fyne.NewMenuItem("Clear", iApp.clearAction),
			fyne.NewMenuItem("Save", iApp.saveAction),
		),
		fyne.NewMenu("Edit",
			addImgsMenuItem,
			delImgsMenuItem,
			dupImgsMenuItem,
			moveUpImgsMenuItem,
			moveDownImgsMenuItem,
			downloadImgsMenuItem,
			fyne.NewMenuItemSeparator(),
			rotateImgsMenuItem,
			cutImgMenuItem,
		),
		fyne.NewMenu("Help",
			fyne.NewMenuItem("Preferences", iApp.showPreferences),
			fyne.NewMenuItem("About", iApp.showAbout),
		),
	)

	iApp.mainWindow.SetMainMenu(menu)
}

func (iApp *ImgpackApp) setupToolbar() {
	addImgsToolbarAction := widget.NewToolbarAction(theme.ContentAddIcon(), iApp.addAction)
	delImgsToolbarAction := widget.NewToolbarAction(theme.DeleteIcon(), iApp.deleteAction)
	dupImgsToolbarAction := widget.NewToolbarAction(theme.ContentCopyIcon(), iApp.dupAction)
	moveUpImgsToolbarAction := widget.NewToolbarAction(theme.MoveUpIcon(), iApp.moveUpAction)
	moveDownImgsToolbarAction := widget.NewToolbarAction(theme.MoveDownIcon(), iApp.moveDownAction)
	downloadImgsToolbarAction := widget.NewToolbarAction(theme.DownloadIcon(), iApp.downloadAction)

	rotateImgsToolbarAction := widget.NewToolbarAction(theme.MediaReplayIcon(), iApp.rotateAction)
	cutImgToolbarAction := widget.NewToolbarAction(theme.ContentCutIcon(), iApp.cutAction)

	iApp.enableOnSelectImageEnables = append(
		iApp.enableOnSelectImageEnables,
		delImgsToolbarAction,
		dupImgsToolbarAction,
		moveUpImgsToolbarAction,
		moveDownImgsToolbarAction,
		downloadImgsToolbarAction,
		rotateImgsToolbarAction,
		cutImgToolbarAction,
	)

	iApp.toolbar = widget.NewToolbar(
		widget.NewToolbarAction(theme.DocumentCreateIcon(), iApp.clearAction),
		widget.NewToolbarAction(theme.DocumentSaveIcon(), iApp.saveAction),
		widget.NewToolbarSeparator(),
		addImgsToolbarAction,
		delImgsToolbarAction,
		dupImgsToolbarAction,
		moveUpImgsToolbarAction,
		moveDownImgsToolbarAction,
		downloadImgsToolbarAction,
		widget.NewToolbarSeparator(),
		rotateImgsToolbarAction,
		cutImgToolbarAction,
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.SettingsIcon(), iApp.showPreferences),
		widget.NewToolbarAction(theme.HelpIcon(), iApp.showAbout),
	)
}

func (iApp *ImgpackApp) setupContent() {
	imgListWidget := widget.NewList(
		func() int {
			return len(iApp.imgs)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Item")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(iApp.imgs[i].Filename)
		},
	)
	imgListWidget.OnSelected = iApp.onSelectImageURI
	iApp.imgListWidget = imgListWidget

	imgShow := canvas.NewImageFromImage(assets.ImgPlaceholder)
	imgShow.FillMode = canvas.ImageFillContain
	iApp.imgShow = imgShow

	hSplit := container.NewHSplit(imgListWidget, imgShow)
	hSplit.SetOffset(0.25)

	stateBar := widget.NewLabel("Ready")
	iApp.stateBar = stateBar

	content := container.NewBorder(iApp.toolbar,
		stateBar, nil, nil, hSplit)

	iApp.mainWindow.SetContent(content)
}

func (iApp *ImgpackApp) Run() {
	iApp.mainWindow.ShowAndRun()
}

func (iApp *ImgpackApp) showPreferences() {
	dlg := dialog.NewCustom("Preference", "OK", preferenceContent(), iApp.mainWindow)
	dlg.Resize(fyne.NewSize(400, 300))
	dlg.Show()
}

func (iApp *ImgpackApp) showAbout() {
	docURL, _ := url.Parse(appURL)
	links := []*widget.Hyperlink{
		widget.NewHyperlink("Github", docURL),
	}

	dialogx.ShowAbout(appDescription, links, iApp.fApp, iApp.mainWindow)
}

func (iApp *ImgpackApp) dropFiles(files []fyne.URI) {
	for _, file := range files {
		imgs, err := imgutil.ReadImgs(file.Path())
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

func (iApp *ImgpackApp) clearSelected() bool {
	iApp.selectedImgIdx = nil
	iApp.imgShow.Resource = nil
	iApp.imgShow.Image = assets.ImgPlaceholder
	iApp.imgShow.Refresh()
	iApp.imgListWidget.UnselectAll()
	iApp.imgListWidget.Refresh()

	for _, action := range iApp.enableOnSelectImageEnables {
		action.Disable()
	}
	return true
}

func (iApp *ImgpackApp) clearAction() {
	if len(iApp.imgs) == 0 {
		return
	}

	dialog.ShowConfirm("Clear all images", "Are you sure to clear all images?",
		func(b bool) {
			if b {
				iApp.imgs = []*imgutil.Image{}
				iApp.clearSelected()
			}
		},
		iApp.mainWindow)
}

func (iApp *ImgpackApp) addAction() {
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

		iApp.readingImagesDlg.Show()
		defer iApp.readingImagesDlg.Hide()

		filepath := f.URI().Path()
		imgs, err := imgutil.ReadImgs(filepath)
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

	dlg.SetFilter(storage.NewExtensionFileFilter(slices.Concat(
		imgutil.SupportedImageExts,
		imgutil.SupportedArchiveExts,
		imgutil.SupportedPDFExts)))
	dlg.Resize(fyne.NewSize(600, 600))
	dlg.Show()
}

func (iApp *ImgpackApp) downloadAction() {
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

		iApp.savingDlg.Show()
		defer iApp.savingDlg.Hide()

		err = imgutil.SaveImg(img, f, getPreferenceJPGQuality())
		if err != nil {
			dialog.ShowError(err, iApp.mainWindow)
			return
		}
		f.Close()

		iApp.stateBar.SetText("Saved successfully")
	}, iApp.mainWindow)

	dlg.SetFileName(img.Filename + ".jpg")
	dlg.SetFilter(storage.NewExtensionFileFilter(imgutil.SupportedImageExts))
	dlg.Resize(fyne.NewSize(600, 600))
	dlg.Show()
}

func (iApp *ImgpackApp) onSelectImageURI(id widget.ListItemID) {
	iApp.selectedImgIdx = &id
	img := iApp.imgs[id]

	for _, action := range iApp.enableOnSelectImageEnables {
		action.Enable()
	}

	bound := img.Img.Bounds()
	imgDesc := fmt.Sprintf("filename: %s, format: %s, size: %dx%d",
		img.Filename, img.Type, bound.Dx(), bound.Dy())
	iApp.stateBar.SetText(imgDesc)

	iApp.imgShow.Resource = nil
	iApp.imgShow.Image = img.Img
	iApp.imgShow.Refresh()
}

func (iApp *ImgpackApp) deleteAction() {
	if iApp.selectedImgIdx == nil {
		return
	}

	idx := *iApp.selectedImgIdx
	iApp.imgs = slices.Delete(iApp.imgs, idx, idx+1)
	iApp.imgListWidget.Refresh()

	if idx >= len(iApp.imgs) {
		iApp.clearSelected()
	} else {
		iApp.onSelectImageURI(idx)
	}
}

func (iApp *ImgpackApp) dupAction() {
	if iApp.selectedImgIdx == nil {
		return
	}

	idx := *iApp.selectedImgIdx
	img := iApp.imgs[idx]

	newImg := img.Clone()

	iApp.imgs = slices.Insert(iApp.imgs, idx+1, newImg)
	iApp.imgListWidget.Refresh()
}

func (iApp *ImgpackApp) moveUpAction() {
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

func (iApp *ImgpackApp) moveDownAction() {
	if iApp.selectedImgIdx == nil {
		return
	}

	idx := *iApp.selectedImgIdx
	if idx == len(iApp.imgs)-1 {
		iApp.imgs = append([]*imgutil.Image{iApp.imgs[len(iApp.imgs)-1]}, iApp.imgs[:len(iApp.imgs)-1]...)
		iApp.imgListWidget.Select(0)
		return
	}

	iApp.imgs[idx], iApp.imgs[idx+1] = iApp.imgs[idx+1], iApp.imgs[idx]
	iApp.imgListWidget.Select(idx + 1)
}

func (iApp *ImgpackApp) saveAction() {
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

		iApp.savingDlg.Show()
		defer iApp.savingDlg.Hide()

		err = imgutil.SaveImgsAsZip(iApp.imgs, f,
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
	dlg.SetFilter(storage.NewExtensionFileFilter(imgutil.SupportedArchiveExts))
	dlg.Resize(fyne.NewSize(600, 600))
	dlg.Show()
}

func (iApp *ImgpackApp) rotateAction() {
	if iApp.selectedImgIdx == nil {
		return
	}

	img := iApp.imgs[*iApp.selectedImgIdx]
	img.Img = imaging.Rotate90(img.Img)
	iApp.imgShow.Image = img.Img
	iApp.imgShow.Refresh()
}

func (iApp *ImgpackApp) cutAction() {
	if iApp.selectedImgIdx == nil {
		return
	}

	idx := *iApp.selectedImgIdx
	img := iApp.imgs[idx]
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

	iApp.imgs = slices.Insert(iApp.imgs, idx+1, newImg)
	iApp.imgListWidget.Refresh()

	iApp.imgShow.Image = img.Img
	iApp.imgShow.Refresh()
}
