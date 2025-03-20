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

	stateBar      *widget.Label
	imgListWidget *widget.List
	imgShow       *canvas.Image

	selectedImgIdx *int

	imgs []*imgutil.Image

	enableOnSelectImageToolbarActions []*widget.ToolbarAction
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

	addImgsToolbarAction := widget.NewToolbarAction(theme.ContentAddIcon(), retApp.toolbarAddAction)
	delImgsToolbarAction := widget.NewToolbarAction(theme.DeleteIcon(), retApp.toolbarDeleteAction)
	dupImgsToolbarAction := widget.NewToolbarAction(theme.ContentCopyIcon(), retApp.toolbarDupAction)
	moveUpImgsToolbarAction := widget.NewToolbarAction(theme.MoveUpIcon(), retApp.toolbarMoveUpAction)
	moveDownImgsToolbarAction := widget.NewToolbarAction(theme.MoveDownIcon(), retApp.toolbarMoveDownAction)
	downloadImgsToolbarAction := widget.NewToolbarAction(theme.DownloadIcon(), retApp.toolbarDownloadAction)
	rotateImgsToolbarAction := widget.NewToolbarAction(theme.MediaReplayIcon(), retApp.toolbarRotateAction)

	retApp.enableOnSelectImageToolbarActions = []*widget.ToolbarAction{
		delImgsToolbarAction,
		dupImgsToolbarAction,
		moveUpImgsToolbarAction,
		moveDownImgsToolbarAction,
		downloadImgsToolbarAction,
		rotateImgsToolbarAction,
	}

	for _, action := range retApp.enableOnSelectImageToolbarActions {
		action.Disable()
	}

	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.DocumentCreateIcon(), retApp.toolbarClearAction),
		widget.NewToolbarAction(theme.DocumentSaveIcon(), retApp.toolbarSaveAction),
		widget.NewToolbarSeparator(),
		addImgsToolbarAction,
		delImgsToolbarAction,
		dupImgsToolbarAction,
		moveUpImgsToolbarAction,
		moveDownImgsToolbarAction,
		downloadImgsToolbarAction,
		widget.NewToolbarSeparator(),
		rotateImgsToolbarAction,
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.SettingsIcon(), func() {
			dlg := dialog.NewCustom("Preference", "OK", preferenceContent(), mainWindow)
			dlg.Resize(fyne.NewSize(400, 300))
			dlg.Show()
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
			o.(*widget.Label).SetText(retApp.imgs[i].Filename)
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

	return retApp
}

func (iApp *ImgpackApp) Run() {
	iApp.mainWindow.ShowAndRun()
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

	for _, action := range iApp.enableOnSelectImageToolbarActions {
		action.Disable()
	}
	return true
}

func (iApp *ImgpackApp) toolbarClearAction() {
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

	for _, action := range iApp.enableOnSelectImageToolbarActions {
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

func (iApp *ImgpackApp) toolbarDeleteAction() {
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
		iApp.imgs = append([]*imgutil.Image{iApp.imgs[len(iApp.imgs)-1]}, iApp.imgs[:len(iApp.imgs)-1]...)
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

func (iApp *ImgpackApp) toolbarRotateAction() {
	if iApp.selectedImgIdx == nil {
		return
	}

	img := iApp.imgs[*iApp.selectedImgIdx]
	img.Img = imaging.Rotate90(img.Img)
	iApp.imgShow.Image = img.Img
	iApp.imgShow.Refresh()
}
