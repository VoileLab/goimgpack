package imgpack

import (
	"fyne.io/fyne/v2"
)

const (
	PreferencePrependDigitKey = "prepend_digit"
	PreferenceJPGQualityKey   = "jpg_quality"
)

func getPreferencePrependDigit() bool {
	return fyne.CurrentApp().Preferences().BoolWithFallback(PreferencePrependDigitKey, true)
}

func setPreferencePrependDigit(value bool) {
	fyne.CurrentApp().Preferences().SetBool(PreferencePrependDigitKey, value)
}

func getPreferenceJPGQuality() int {
	return fyne.CurrentApp().Preferences().IntWithFallback(PreferenceJPGQualityKey, 100)
}

func setPreferenceJPGQuality(value int) {
	fyne.CurrentApp().Preferences().SetInt(PreferenceJPGQualityKey, value)
}
