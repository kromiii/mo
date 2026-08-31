package desktop

import (
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// CreateAppMenu builds the macOS native menu bar.
func CreateAppMenu(app *App) *menu.Menu {
	appMenu := menu.NewMenu()

	// macOS Application menu
	appMenu.Append(menu.AppMenu())

	// File Menu
	fileMenu := appMenu.AddSubmenu("File")
	fileMenu.AddText("Open File...", keys.CmdOrCtrl("o"), func(_ *menu.CallbackData) {
		app.SelectAndOpenFile()
	})
	fileMenu.AddText("Open Folder...", keys.Combo("o", keys.CmdOrCtrlKey, keys.ShiftKey), func(_ *menu.CallbackData) {
		app.SelectAndOpenDirectory()
	})
	fileMenu.AddSeparator()
	fileMenu.AddText("Close Window", keys.CmdOrCtrl("w"), func(cd *menu.CallbackData) {
		if app.ctx != nil {
			runtime.WindowHide(app.ctx)
		}
	})

	// Edit Menu (Standard OS Undo, Cut, Copy, Paste, Select All)
	appMenu.Append(menu.EditMenu())

	// View Menu
	viewMenu := appMenu.AddSubmenu("View")
	viewMenu.AddText("Reload", keys.CmdOrCtrl("r"), func(_ *menu.CallbackData) {
		if app.ctx != nil {
			runtime.WindowReload(app.ctx)
		}
	})
	viewMenu.AddSeparator()
	viewMenu.AddText("Actual Size", keys.CmdOrCtrl("0"), func(_ *menu.CallbackData) {
		if app.ctx != nil {
			runtime.WindowSetSize(app.ctx, 1200, 800)
			runtime.WindowCenter(app.ctx)
		}
	})
	viewMenu.AddText("Toggle Full Screen", keys.Combo("f", keys.CmdOrCtrlKey, keys.ControlKey), func(_ *menu.CallbackData) {
		if app.ctx != nil {
			runtime.WindowToggleMaximise(app.ctx)
		}
	})

	// Window Menu
	appMenu.Append(menu.WindowMenu())

	return appMenu
}
