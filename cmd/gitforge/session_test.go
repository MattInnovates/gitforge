package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoadSession(t *testing.T) {
	tempRoot := t.TempDir()
	configRoot := filepath.Join(tempRoot, "config")
	if err := os.Setenv("XDG_CONFIG_HOME", configRoot); err != nil {
		t.Fatalf("set XDG_CONFIG_HOME: %v", err)
	}
	if err := os.Setenv("APPDATA", configRoot); err != nil {
		t.Fatalf("set APPDATA: %v", err)
	}
	if err := os.Setenv("LOCALAPPDATA", configRoot); err != nil {
		t.Fatalf("set LOCALAPPDATA: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Unsetenv("XDG_CONFIG_HOME")
		_ = os.Unsetenv("APPDATA")
		_ = os.Unsetenv("LOCALAPPDATA")
	})

	want := session{
		Server:   "http://localhost:3000",
		Username: "matt",
		Token:    "abc123",
	}

	if err := saveSession(want); err != nil {
		t.Fatalf("saveSession: %v", err)
	}

	got, err := loadSession()
	if err != nil {
		t.Fatalf("loadSession: %v", err)
	}

	if got.Server != want.Server || got.Username != want.Username || got.Token != want.Token {
		t.Fatalf("session mismatch got=%+v want=%+v", got, want)
	}
}
