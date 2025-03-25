package imgpack

import "fyne.io/fyne/v2"

type Enablable interface {
	Enable()
	Disable()
}

type EnablableWrapMenuItem struct {
	*fyne.MenuItem
}

func (e *EnablableWrapMenuItem) Enable() {
	e.Disabled = false
}

func (e *EnablableWrapMenuItem) Disable() {
	e.Disabled = true
}
