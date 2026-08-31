package desktop

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/k1LoW/donegroup"
	"github.com/k1LoW/mo/internal/backup"
	"github.com/k1LoW/mo/internal/server"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App represents the desktop application state and lifecycle.
type App struct {
	ctx          context.Context
	cancel       context.CancelFunc
	state        *server.State
	httpServer   *http.Server
	port         int
	initialFiles []string
	mu           sync.Mutex
}

// NewApp creates a new desktop App instance.
func NewApp(port int, initialFiles []string) *App {
	return &App{
		port:         port,
		initialFiles: initialFiles,
	}
}

// Startup is called when the Wails application starts up.
func (a *App) Startup(ctx context.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.ctx, a.cancel = donegroup.WithCancel(ctx)
	a.state = server.NewState(a.ctx)

	// Enable session backup
	a.state.EnableBackup(a.ctx, func(data server.RestoreData) {
		if err := backup.Save(a.port, data); err != nil {
			slog.Warn("failed to save backup", "error", err)
		}
	})

	// Restore previous session
	var rd server.RestoreData
	if err := backup.Load(a.port, &rd); err == nil {
		for group, files := range rd.Groups {
			for _, f := range files {
				if f != "" {
					_, _ = a.state.AddFile(f, group)
				}
			}
		}
		for group, patterns := range rd.Patterns {
			for _, p := range patterns {
				if p != "" {
					_, _ = a.state.AddPattern(p, group)
				}
			}
		}
		for _, uf := range rd.UploadedFiles {
			a.state.AddUploadedFile(uf.Name, uf.Content, uf.Group)
		}
	}

	// Add any initial files passed via CLI or OS open event
	for _, f := range a.initialFiles {
		_ = a.openPathLocked(f)
	}

	// Start background HTTP listener on 127.0.0.1:port to accept external CLI commands
	handler := server.NewHandler(a.state)
	addr := fmt.Sprintf("127.0.0.1:%d", a.port)
	a.httpServer = &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	ln, err := net.Listen("tcp", addr)
	if err == nil {
		go func() {
			slog.Info("desktop background server listening", "addr", addr)
			if err := a.httpServer.Serve(ln); err != nil && err != http.ErrServerClosed {
				slog.Warn("desktop http server error", "error", err)
			}
		}()
	} else {
		slog.Warn("desktop server could not bind external port (another server may be running)", "addr", addr, "error", err)
	}
}

// Shutdown is called when the Wails application closes.
func (a *App) Shutdown(ctx context.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.httpServer != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = a.httpServer.Shutdown(shutdownCtx)
	}
	if a.state != nil {
		a.state.CloseAllSubscribers()
	}
	if a.cancel != nil {
		a.cancel()
	}
	if a.ctx != nil {
		_ = donegroup.WaitWithTimeout(a.ctx, 2*time.Second)
	}
}

// GetState returns the current server state.
func (a *App) GetState() *server.State {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.state
}

// OpenFile adds a file or directory to the default group.
func (a *App) OpenFile(path string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.openPathLocked(path)
}

func (a *App) openPathLocked(path string) error {
	if a.state == nil {
		return fmt.Errorf("state not initialized")
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}

	fi, err := os.Stat(abs)
	if err == nil && fi.IsDir() {
		_, err = a.state.AddPattern(filepath.Join(abs, "*.md"), "default")
		return err
	}

	_, err = a.state.AddFile(abs, "default")
	return err
}

// SelectAndOpenFile opens a native file dialog to pick markdown files.
func (a *App) SelectAndOpenFile() {
	if a.ctx == nil {
		return
	}
	selection, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Open Markdown File",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Markdown Files (*.md, *.markdown, *.mdx)",
				Pattern:     "*.md;*.markdown;*.mdx;*.mdown;*.mkd",
			},
			{
				DisplayName: "All Files (*.*)",
				Pattern:     "*.*",
			},
		},
	})
	if err == nil && selection != "" {
		_ = a.OpenFile(selection)
	}
}

// SelectAndOpenDirectory opens a native folder dialog.
func (a *App) SelectAndOpenDirectory() {
	if a.ctx == nil {
		return
	}
	selection, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Open Markdown Directory",
	})
	if err == nil && selection != "" {
		_ = a.OpenFile(selection)
	}
}
