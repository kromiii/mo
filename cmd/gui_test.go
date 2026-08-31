package cmd

import (
	"testing"
)

func TestGUIAvailability(t *testing.T) {
	if !isGUIAvailable() {
		t.Log("GUI is not available in this build configuration")
		return
	}
	if guiCmd == nil {
		t.Fatal("expected non-nil guiCmd")
	}
	if guiCmd.Name() != "gui" {
		t.Errorf("expected command name 'gui', got %q", guiCmd.Name())
	}
}
