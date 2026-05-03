package main

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestParseGitCommand(t *testing.T) {
	tests := []struct {
		name        string
		in          string
		wantService string
		wantRepo    string
		wantErr     bool
	}{
		{name: "upload-pack single quoted", in: "git-upload-pack 'owner/repo.git'", wantService: "git-upload-pack", wantRepo: "owner/repo.git"},
		{name: "receive-pack double quoted", in: `git-receive-pack "owner/repo.git"`, wantService: "git-receive-pack", wantRepo: "owner/repo.git"},
		{name: "unsupported service", in: "git-shell owner/repo.git", wantErr: true},
		{name: "bad format", in: "git-upload-pack", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotService, gotRepo, err := parseGitCommand(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotService != tt.wantService {
				t.Fatalf("service mismatch: got %q want %q", gotService, tt.wantService)
			}
			if gotRepo != tt.wantRepo {
				t.Fatalf("repo mismatch: got %q want %q", gotRepo, tt.wantRepo)
			}
		})
	}
}

func TestResolveRepoPath(t *testing.T) {
	root := t.TempDir()

	got, err := resolveRepoPath(root, "owner/repo.git")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := filepath.Join(root, "owner", "repo.git")
	if got != want {
		t.Fatalf("path mismatch: got %q want %q", got, want)
	}
}

func TestResolveRepoPathRejectEscape(t *testing.T) {
	root := t.TempDir()

	_, err := resolveRepoPath(root, "../outside.git")
	if err == nil {
		t.Fatal("expected path escape error")
	}
}

func TestResolveRepoPathRejectAbsolute(t *testing.T) {
	root := t.TempDir()
	absolute := "/outside.git"
	if runtime.GOOS == "windows" {
		absolute = `C:\outside.git`
	}

	_, err := resolveRepoPath(root, absolute)
	if err == nil {
		t.Fatal("expected absolute path error")
	}
}
