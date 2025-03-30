package imgpack

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

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

type EnablableWrapToolbarItem struct {
	item *widget.ToolbarAction

	enabledIcon  fyne.Resource
	disabledIcon fyne.Resource
}

func NewEnablableWrapToolbarItem(enabledIcon, disabledIcon fyne.Resource,
	onActivated func()) *EnablableWrapToolbarItem {

	return &EnablableWrapToolbarItem{
		item:         widget.NewToolbarAction(enabledIcon, onActivated),
		enabledIcon:  enabledIcon,
		disabledIcon: disabledIcon,
	}
}

func (e *EnablableWrapToolbarItem) ToolbarItem() *widget.ToolbarAction {
	return e.item
}

func (e *EnablableWrapToolbarItem) Enable() {
	e.item.SetIcon(e.enabledIcon)
	e.item.Enable()
}

func (e *EnablableWrapToolbarItem) Disable() {
	e.item.SetIcon(e.disabledIcon)
	e.item.Disable()
}
