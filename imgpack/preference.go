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

// GetPreferenceScale returns the scale factor of the application.
func GetPreferenceScale() float64 {
	conf, err := getConf()
	if err != nil {
		return 1
	}

	s := conf.Scale
	if s == nil || *s < 0.5 || *s > 4 {
		return 1
	}

	return *conf.Scale
}

func setPreferenceScale(value float64) {
	if value < 0.5 || value > 4 {
		return
	}

	c, err := getConf()
	if err != nil {
		c = &conf{}
	}

	c.Scale = &value
	setConf(c)
}
