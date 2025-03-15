package imgpackgui

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var supportedImageExts = []string{".png", ".jpg", ".jpeg", ".webp"}

type ImgpackApp struct {
	fApp   fyne.App
	window fyne.Window

	stateBar      *widget.Label
	imgListWidget *widget.List

	imgURIs []string
}

func NewImgpackApp() *ImgpackApp {
	fApp := app.NewWithID("com.mukyu.voile.imgpack")
	window := fApp.NewWindow("Image Packer")
	window.Resize(fyne.NewSize(800, 600))
	window.CenterOnScreen()

	retApp := &ImgpackApp{
		fApp:   fApp,
		window: window,

		imgURIs: []string{},
	}

	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.ContentAddIcon(), retApp.toolbarAddAction),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.ContentCutIcon(), func() {}),
		widget.NewToolbarAction(theme.ContentCopyIcon(), func() {}),
		widget.NewToolbarAction(theme.ContentPasteIcon(), func() {}),
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.HelpIcon(), func() {
			// show messagebox
			dialog := dialog.NewInformation("About", "Image Packer v0.1", window)
			dialog.Show()
		}),
	)

	imgListWidget := widget.NewList(
		func() int {
			return len(retApp.imgURIs)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Item")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(retApp.imgURIs[i])
		},
	)
	imgListWidget.OnSelected = retApp.onSelectImageURI
	retApp.imgListWidget = imgListWidget

	msgLabel := widget.NewEntry()
	msgLabel.SetPlaceHolder("Enter text...")
	msgLabel.MultiLine = true

	hSplit := container.NewHSplit(imgListWidget, msgLabel)
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
	dialog := dialog.NewFileOpen(func(f fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, app.window)
			return
		}

		if f == nil {
			log.Println("Cancelled")
			return
		}

		uri := f.URI()
		if uri == nil {
			log.Println("URI is nil")
			return
		}

		app.imgURIs = append(app.imgURIs, uri.String())
		app.imgListWidget.Refresh()
	}, app.window)
	dialog.SetFilter(storage.NewExtensionFileFilter(supportedImageExts))
	dialog.Show()
}

func (app *ImgpackApp) onSelectImageURI(id widget.ListItemID) {
	app.stateBar.SetText(app.imgURIs[id])
}
