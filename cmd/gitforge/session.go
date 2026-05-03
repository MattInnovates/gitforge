package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type session struct {
	Server   string `json:"server"`
	Username string `json:"username"`
	Token    string `json:"token"`
}

func sessionFilePath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}
	return filepath.Join(configDir, "gitforge", "session.json"), nil
}

func saveSession(s session) error {
	path, err := sessionFilePath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create session directory: %w", err)
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write session file: %w", err)
	}

	return nil
}

func loadSession() (session, error) {
	path, err := sessionFilePath()
	if err != nil {
		return session{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return session{}, fmt.Errorf("read session file: %w", err)
	}

	var s session
	if err := json.Unmarshal(data, &s); err != nil {
		return session{}, fmt.Errorf("decode session file: %w", err)
	}
	if s.Server == "" || s.Token == "" {
		return session{}, fmt.Errorf("session file is incomplete")
	}

	return s, nil
}
