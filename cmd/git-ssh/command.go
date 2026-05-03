package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var allowedGitServices = map[string]string{
	"git-upload-pack":  "upload-pack",
	"git-receive-pack": "receive-pack",
}

func parseGitCommand(raw string) (service string, repo string, err error) {
	trimmed := strings.TrimSpace(raw)
	parts := strings.Fields(trimmed)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("unsupported SSH command format")
	}

	service = parts[0]
	if _, ok := allowedGitServices[service]; !ok {
		return "", "", fmt.Errorf("unsupported SSH service %q", service)
	}

	repo = strings.Trim(parts[1], "'\"")
	repo = strings.TrimLeft(repo, "/\\")
	if repo == "" {
		return "", "", fmt.Errorf("repository path is empty")
	}

	return service, repo, nil
}

func resolveRepoPath(repoRoot, repo string) (string, error) {
	rootAbs, err := filepath.Abs(repoRoot)
	if err != nil {
		return "", fmt.Errorf("resolve repo root: %w", err)
	}

	repoClean := filepath.Clean(filepath.FromSlash(repo))
	if filepath.IsAbs(repoClean) {
		return "", fmt.Errorf("absolute repository path is not allowed")
	}

	repoAbs := filepath.Clean(filepath.Join(rootAbs, repoClean))
	rel, err := filepath.Rel(rootAbs, repoAbs)
	if err != nil {
		return "", fmt.Errorf("calculate relative repo path: %w", err)
	}

	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("repository path escapes repo root")
	}

	return repoAbs, nil
}
