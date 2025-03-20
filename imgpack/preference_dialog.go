package imgpack

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func preferenceContent() fyne.CanvasObject {
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

	appScaleLabel := widget.NewLabel(
		fmt.Sprintf("App Scale: %.2f", GetPreferenceScale()))

	appScaleSlider := widget.NewSlider(0.5, 4)
	appScaleSlider.Step = 0.1
	appScaleSlider.Value = GetPreferenceScale()
	appScaleSlider.OnChanged = func(v float64) {
		setPreferenceScale(v)
		appScaleLabel.SetText(
			fmt.Sprintf("App Scale: %.2f\n(Take effect after restart)", v))
	}

	return container.New(layout.NewFormLayout(),
		widget.NewLabel("Add digit to filename"),
		addDigitCheck,
		jpgQualitySliderLabel,
		jpgQualitySlider,
		appScaleLabel,
		appScaleSlider,
	)
}
