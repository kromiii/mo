package desktop

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestNewApp(t *testing.T) {
	app := NewApp(16275, []string{"testdata/basic.md"})
	if app == nil {
		t.Fatal("expected non-nil App")
	}
	if app.port != 16275 {
		t.Errorf("expected port 16275, got %d", app.port)
	}
	if len(app.initialFiles) != 1 || app.initialFiles[0] != "testdata/basic.md" {
		t.Errorf("unexpected initialFiles: %v", app.initialFiles)
	}
}

func TestAppOpenFileAndDir(t *testing.T) {
	app := NewApp(16276, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app.Startup(ctx)
	defer app.Shutdown(ctx)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "hello.md")
	if err := os.WriteFile(testFile, []byte("# Hello Desktop"), 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	if err := app.OpenFile(testFile); err != nil {
		t.Fatalf("failed to open file: %v", err)
	}

	st := app.GetState()
	if st == nil {
		t.Fatal("expected non-nil state")
	}

	groups := st.Groups()
	if len(groups) == 0 {
		t.Fatal("expected at least one group")
	}

	// Test directory open
	if err := app.OpenFile(tmpDir); err != nil {
		t.Fatalf("failed to open directory: %v", err)
	}
}
