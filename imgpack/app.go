package imgpack

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	dialogx "fyne.io/x/fyne/dialog"

	"github.com/VoileLab/goimgpack/imgpack/assets"
	"github.com/VoileLab/goimgpack/imgpack/imgstable"
	"github.com/VoileLab/goimgpack/internal/imgutil"
)

const appURL = "https://github.com/VoileLab/goimgpack"

var appSize = fyne.NewSize(1000, 800)

type ImgpackApp struct {
	fApp       fyne.App
	mainWindow fyne.Window

	toolbar *widget.Toolbar

	stateBar      *widget.Label
	imgListWidget *widget.List
	imgShow       *canvas.Image

	opTable *imgstable.ImgsTable

	enableOnSelectImageEnables []Enablable

	// dialogs
	readingImagesDlg dialog.Dialog
	savingDlg        dialog.Dialog
	aboutDlg         dialog.Dialog
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

		opTable: imgstable.New(),
	}

	retApp.opTable.SetOnSelectIndexChange(func() {
		if !retApp.opTable.IsSelected() {
			retApp.imgShow.Resource = nil
			retApp.imgShow.Image = assets.ImgPlaceholder
			retApp.imgShow.Refresh()
			retApp.imgListWidget.UnselectAll()
			retApp.imgListWidget.Refresh()

			for _, action := range retApp.enableOnSelectImageEnables {
				action.Disable()
			}

			return
		}

		img := retApp.opTable.GetSelectedImg()

		for _, action := range retApp.enableOnSelectImageEnables {
			action.Enable()
		}

		bound := img.Img.Bounds()
		imgDesc := fmt.Sprintf("filename: %s, format: %s, size: %dx%d",
			img.Filename, img.Type, bound.Dx(), bound.Dy())

		retApp.stateBar.SetText(imgDesc)

		retApp.imgShow.Resource = nil
		retApp.imgShow.Image = img.Img
		retApp.imgShow.Refresh()

		retApp.imgListWidget.Select(retApp.opTable.GetSelectedIdx())
	})

	retApp.opTable.SetOnSelectImageChange(func() {
		retApp.imgShow.Image = retApp.opTable.GetSelectedImg().Img
		retApp.imgShow.Refresh()
	})

	retApp.opTable.SetOnListChange(func() {
		retApp.imgListWidget.Refresh()
	})

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

	docURL, _ := url.Parse(appURL)
	links := []*widget.Hyperlink{
		widget.NewHyperlink("Github", docURL),
	}

	aboutDlg := dialogx.NewAbout(
		assets.AppDescription, links, iApp.fApp, iApp.mainWindow)
	aboutDlg.Resize(fyne.NewSize(500, 400))

	iApp.aboutDlg = aboutDlg
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
				Label:  "Save As Archive",
				Icon:   theme.DocumentSaveIcon(),
				Action: iApp.saveArchiveAction,
			},
			&fyne.MenuItem{
				Label:  "Save As PDF",
				Icon:   theme.DocumentSaveIcon(),
				Action: iApp.savePDFAction,
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
		widget.NewToolbarAction(theme.DocumentSaveIcon(), iApp.saveArchiveAction),
		widget.NewToolbarAction(assets.AsPdfIcon, iApp.savePDFAction),
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
			return iApp.opTable.Len()
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Item")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(iApp.opTable.Get(i).Filename)
		},
	)

	imgListWidget.OnSelected = func(id widget.ListItemID) {
		iApp.opTable.Select(int(id))
	}

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
	iApp.aboutDlg.Show()
}

func (iApp *ImgpackApp) dropFiles(files []fyne.URI) {
	iApp.readingImagesDlg.Show()
	defer iApp.readingImagesDlg.Hide()

	accImgs := []*imgutil.Image{}

	for _, file := range files {
		fileStat, err := os.Stat(file.Path())
		if err != nil {
			dialog.ShowError(err, iApp.mainWindow)
			continue
		}

		if fileStat.IsDir() {
			imgs, err := imgutil.ReadImgsInDir(file.Path())
			if err != nil {
				dialog.ShowError(err, iApp.mainWindow)
				continue
			}

			accImgs = append(accImgs, imgs...)
			continue
		}

		f, err := os.Open(file.Path())
		if err != nil {
			dialog.ShowError(err, iApp.mainWindow)
			continue
		}

		imgs, err := imgutil.ReadImgsInFile(f, path.Base(file.Path()))
		if err != nil {
			dialog.ShowError(err, iApp.mainWindow)
			continue
		}

		accImgs = append(accImgs, imgs...)

	}

	iApp.opTable.Insert(accImgs...)
}

func (iApp *ImgpackApp) onTabKey(e *fyne.KeyEvent) {
	if !iApp.opTable.IsSelected() {
		return
	}

	idx := iApp.opTable.GetSelectedIdx()

	switch e.Name {
	case fyne.KeyUp:
		if idx > 0 {
			iApp.opTable.Select(idx - 1)
		} else {
			iApp.opTable.Select(iApp.opTable.Len() - 1)
		}
	case fyne.KeyDown:
		if idx < iApp.opTable.Len()-1 {
			iApp.opTable.Select(idx + 1)
		} else {
			iApp.opTable.Select(0)
		}
	}
}

func (iApp *ImgpackApp) clearAction() {
	if iApp.opTable.Len() == 0 {
		return
	}

	dialog.ShowConfirm("Clear all images", "Are you sure to clear all images?",
		func(b bool) {
			if b {
				iApp.opTable.Clear()
			}
		},
		iApp.mainWindow)
}

func (iApp *ImgpackApp) addAction() {
	openImgsFile(func(f fyne.URIReadCloser) {
		iApp.readingImagesDlg.Show()
		defer iApp.readingImagesDlg.Hide()

		if f.URI() == nil {
			log.Println("URI is nil")
			return
		}

		filepath := f.URI().Path()
		imgs, err := imgutil.ReadImgsInFile(f, path.Base(filepath))
		if err != nil {
			dialog.ShowError(err, iApp.mainWindow)
			return
		}

		iApp.opTable.Insert(imgs...)
	}, iApp.mainWindow)
}

func (iApp *ImgpackApp) downloadAction() {
	if !iApp.opTable.IsSelected() {
		return
	}

	img := iApp.opTable.GetSelectedImg()
	saveImgFile(img.Filename+".jpg", func(f fyne.URIWriteCloser) {
		iApp.savingDlg.Show()
		defer iApp.savingDlg.Hide()

		err := imgutil.SaveImg(img, f, getPreferenceJPGQuality())
		if err != nil {
			dialog.ShowError(err, iApp.mainWindow)
			return
		}
		f.Close()

		iApp.stateBar.SetText("Saved successfully")
	}, iApp.mainWindow)
}

func (iApp *ImgpackApp) deleteAction() {
	iApp.opTable.Delete()
}

func (iApp *ImgpackApp) dupAction() {
	iApp.opTable.Duplicate()
}

func (iApp *ImgpackApp) moveUpAction() {
	iApp.opTable.MoveUp()
}

func (iApp *ImgpackApp) moveDownAction() {
	iApp.opTable.MoveDown()
}

func (iApp *ImgpackApp) saveArchiveAction() {
	if iApp.opTable.Len() == 0 {
		iApp.stateBar.SetText("No image to save")
		return
	}

	saveArchiveFile("output.cbz", func(f fyne.URIWriteCloser) {
		iApp.savingDlg.Show()
		defer iApp.savingDlg.Hide()

		err := imgutil.SaveImgsAsZip(
			iApp.opTable.GetImgs(), f,
			getPreferencePrependDigit(),
			getPreferenceJPGQuality())
		if err != nil {
			dialog.ShowError(err, iApp.mainWindow)
			return
		}
		f.Close()

		iApp.stateBar.SetText("Saved successfully")
	}, iApp.mainWindow)
}

func (iApp *ImgpackApp) savePDFAction() {
	if iApp.opTable.Len() == 0 {
		iApp.stateBar.SetText("No image to save")
		return
	}

	savePDFFile("output.pdf", func(f fyne.URIWriteCloser) {
		iApp.savingDlg.Show()
		defer iApp.savingDlg.Hide()

		err := imgutil.SaveImgsAsPDF(
			iApp.opTable.GetImgs(), f, getPreferenceJPGQuality())
		if err != nil {
			dialog.ShowError(err, iApp.mainWindow)
			return
		}
		f.Close()

		iApp.stateBar.SetText("Saved successfully")
	}, iApp.mainWindow)
}

func (iApp *ImgpackApp) rotateAction() {
	iApp.opTable.Rotate()
}

func (iApp *ImgpackApp) cutAction() {
	iApp.opTable.Cut()
}
