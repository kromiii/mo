package desktop

import (
	"fmt"
	"io/fs"
	"net/http"

	"github.com/k1LoW/mo/internal/server"
	"github.com/k1LoW/mo/internal/static"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
)

// Run initializes and runs the Wails desktop application.
func Run(port int, initialFiles []string) error {
	app := NewApp(port, initialFiles)

	distFS, err := fs.Sub(static.Frontend, "dist")
	if err != nil {
		return fmt.Errorf("failed to load static frontend: %w", err)
	}

	appMenu := CreateAppMenu(app)

	return wails.Run(&options.App{
		Title:     "mo",
		Width:     1200,
		Height:    800,
		MinWidth:  600,
		MinHeight: 400,
		Menu:      appMenu,
		AssetServer: &assetserver.Options{
			Assets: distFS,
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				st := app.GetState()
				if st != nil {
					h := server.NewHandler(st)
					h.ServeHTTP(w, r)
					return
				}
				http.NotFound(w, r)
			}),
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 1},
		OnStartup:        app.Startup,
		OnShutdown:       app.Shutdown,
		Bind: []interface{}{
			app,
		},
		DragAndDrop: &options.DragAndDrop{
			EnableFileDrop:     true,
			DisableWebViewDrop: false,
		},
		Mac: &mac.Options{
			TitleBar:   mac.TitleBarDefault(),
			Appearance: mac.DefaultAppearance,
			About: &mac.AboutInfo{
				Title:   "mo",
				Message: "Markdown viewer",
			},
			OnFileOpen: func(filePath string) {
				_ = app.OpenFile(filePath)
			},
		},
	})
}
