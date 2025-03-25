package imgpack

import (
	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"github.com/VoileLab/goimgpack/internal/imgutil"
)

func openImgsFile(cb func(fyne.URIReadCloser), w fyne.Window) {
	dlg := dialog.NewFileOpen(func(f fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, w)
			return
		}

		if f == nil {
			return
		}

		cb(f)
	}, w)

	dlg.SetFilter(storage.NewExtensionFileFilter(slices.Concat(
		imgutil.SupportedImageExts,
		imgutil.SupportedArchiveExts,
		imgutil.SupportedPDFExts)))
	dlg.Resize(fyne.NewSize(600, 600))
	dlg.Show()
}

func saveImgFile(defaultName string, cb func(fyne.URIWriteCloser), w fyne.Window) {

	dlg := dialog.NewFileSave(func(f fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, w)
			return
		}

		if f == nil {
			return
		}

		cb(f)
	}, w)

	dlg.SetFileName(defaultName)
	dlg.SetFilter(storage.NewExtensionFileFilter(imgutil.SupportedImageExts))
	dlg.Resize(fyne.NewSize(600, 600))
	dlg.Show()
}

func saveArchiveFile(defaultName string, cb func(fyne.URIWriteCloser), w fyne.Window) {
	dlg := dialog.NewFileSave(func(f fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, w)
			return
		}

		if f == nil {
			return
		}

		cb(f)
	}, w)

	dlg.SetFileName(defaultName)
	dlg.SetFilter(storage.NewExtensionFileFilter(imgutil.SupportedArchiveExts))
	dlg.Resize(fyne.NewSize(600, 600))
	dlg.Show()
}
