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

	opTable *OPTable

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

		opTable: &OPTable{},
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
	addImgsMenuItem := &fyne.MenuItem{
		Label:  "Add",
		Action: iApp.addAction,
		Icon:   theme.ContentAddIcon(),
	}

	delImgsMenuItem := &fyne.MenuItem{
		Label:  "Delete",
		Action: iApp.deleteAction,
		Icon:   theme.DeleteIcon(),
	}

	dupImgsMenuItem := &fyne.MenuItem{
		Label:  "Duplicate",
		Action: iApp.dupAction,
		Icon:   theme.ContentCopyIcon(),
	}

	moveUpImgsMenuItem := &fyne.MenuItem{
		Label:  "Move Up",
		Action: iApp.moveUpAction,
		Icon:   theme.MoveUpIcon(),
	}

	moveDownImgsMenuItem := &fyne.MenuItem{
		Label:  "Move Down",
		Action: iApp.moveDownAction,
		Icon:   theme.MoveDownIcon(),
	}

	downloadImgsMenuItem := &fyne.MenuItem{
		Label:  "Download",
		Action: iApp.downloadAction,
		Icon:   theme.DownloadIcon(),
	}

	rotateImgsMenuItem := &fyne.MenuItem{
		Label:  "Rotate",
		Action: iApp.rotateAction,
		Icon:   theme.MediaReplayIcon(),
	}

	cutImgMenuItem := &fyne.MenuItem{
		Label:  "Cut",
		Action: iApp.cutAction,
		Icon:   theme.ContentCutIcon(),
	}

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
			&fyne.MenuItem{
				Label:  "Clear",
				Icon:   theme.DocumentCreateIcon(),
				Action: iApp.clearAction,
			},
			&fyne.MenuItem{
				Label:  "Save",
				Icon:   theme.DocumentSaveIcon(),
				Action: iApp.saveAction,
			},
			fyne.NewMenuItemSeparator(),
			&fyne.MenuItem{
				Label:  "Quit",
				Icon:   theme.CancelIcon(),
				Action: iApp.fApp.Quit,
			},
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
			&fyne.MenuItem{
				Label:  "Preferences",
				Icon:   theme.SettingsIcon(),
				Action: iApp.showPreferences,
			},
			&fyne.MenuItem{
				Label:  "About",
				Icon:   theme.HelpIcon(),
				Action: iApp.showAbout,
			},
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
			return len(iApp.opTable.imgs)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Item")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(iApp.opTable.imgs[i].Filename)
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

		iApp.opTable.Insert(imgs...)
	}

	iApp.imgListWidget.Refresh()
}

func (iApp *ImgpackApp) onTabKey(e *fyne.KeyEvent) {
	if iApp.opTable.selIdx == nil {
		return
	}

	idx := *iApp.opTable.selIdx

	switch e.Name {
	case fyne.KeyUp:
		if idx > 0 {
			iApp.imgListWidget.Select(idx - 1)
		} else {
			iApp.imgListWidget.Select(len(iApp.opTable.imgs) - 1)
		}
	case fyne.KeyDown:
		if idx < len(iApp.opTable.imgs)-1 {
			iApp.imgListWidget.Select(idx + 1)
		} else {
			iApp.imgListWidget.Select(0)
		}
	}
}

func (iApp *ImgpackApp) clearSelected() bool {
	iApp.opTable.Unselect()
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
	if len(iApp.opTable.imgs) == 0 {
		return
	}

	dialog.ShowConfirm("Clear all images", "Are you sure to clear all images?",
		func(b bool) {
			if b {
				iApp.opTable.Clear()
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

		iApp.opTable.Insert(imgs...)
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
	if !iApp.opTable.IsSelected() {
		return
	}

	img := iApp.opTable.GetSelected()
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
	iApp.opTable.Select(int(id))
	img := iApp.opTable.GetSelected()

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
	if !iApp.opTable.IsSelected() {
		return
	}

	idx := *iApp.opTable.selIdx
	iApp.opTable.Delete()
	iApp.imgListWidget.Refresh()

	if idx >= len(iApp.opTable.imgs) {
		iApp.clearSelected()
	} else {
		iApp.onSelectImageURI(idx)
	}
}

func (iApp *ImgpackApp) dupAction() {
	iApp.opTable.Duplicate()
	iApp.imgListWidget.Refresh()
}

func (iApp *ImgpackApp) moveUpAction() {
	if !iApp.opTable.IsSelected() {
		return
	}

	iApp.opTable.MoveUp()
	iApp.imgListWidget.Select(*iApp.opTable.selIdx)
}

func (iApp *ImgpackApp) moveDownAction() {
	if !iApp.opTable.IsSelected() {
		return
	}

	iApp.opTable.MoveDown()
	iApp.imgListWidget.Select(*iApp.opTable.selIdx)
}

func (iApp *ImgpackApp) saveAction() {
	if len(iApp.opTable.imgs) == 0 {
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

		err = imgutil.SaveImgsAsZip(iApp.opTable.imgs, f,
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
	if !iApp.opTable.IsSelected() {
		return
	}

	iApp.opTable.Rotate()

	iApp.imgShow.Image = iApp.opTable.GetSelected().Img
	iApp.imgShow.Refresh()
}

func (iApp *ImgpackApp) cutAction() {
	if !iApp.opTable.IsSelected() {
		return
	}

	iApp.opTable.Cut()

	iApp.imgListWidget.Refresh()
	iApp.imgShow.Image = iApp.opTable.GetSelected().Img
	iApp.imgShow.Refresh()
}
