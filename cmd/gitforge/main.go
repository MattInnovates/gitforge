package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	var err error
	switch cmd {
	case "help", "-h", "--help":
		printUsage()
		return
	case "login":
		err = runLogin(args)
	case "me":
		err = runMe(args)
	case "user-create":
		err = runUserCreate(args)
	case "repo-create":
		err = runRepoCreate(args)
	case "init", "add", "commit", "status", "log", "clone", "push", "pull", "branch", "checkout":
		err = runGit(append([]string{cmd}, args...))
	default:
		err = fmt.Errorf("unknown command %q", cmd)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("GitForge CLI")
	fmt.Println("\nCore commands:")
	fmt.Println("  gitforge login --server <url> --username <name> --password <pwd>")
	fmt.Println("  gitforge me [--server <url>]")
	fmt.Println("  gitforge user-create --server <url> --username <name> --email <email> --password <pwd>")
	fmt.Println("  gitforge repo-create --name <repo> [--owner <name>] [--description <text>] [--visibility private|public] [--server <url>]")
	fmt.Println("\nGit passthrough:")
	fmt.Println("  gitforge init|add|commit|status|log|clone|push|pull|branch|checkout ...")
}

func runLogin(args []string) error {
	fs := flag.NewFlagSet("login", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	server := fs.String("server", "", "GitForge base URL")
	username := fs.String("username", "", "username")
	password := fs.String("password", "", "password")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *server == "" || *username == "" || *password == "" {
		return fmt.Errorf("server, username, and password are required")
	}

	serverURL := strings.TrimRight(*server, "/")
	respBody, err := requestJSON(http.MethodPost, serverURL+"/api/auth/login", map[string]string{
		"username": *username,
		"password": *password,
	}, "")
	if err != nil {
		return err
	}

	var out map[string]string
	if err := json.Unmarshal(respBody, &out); err != nil {
		return fmt.Errorf("decode login response: %w", err)
	}
	token := out["token"]
	if token == "" {
		return fmt.Errorf("login response missing token")
	}

	if err := saveSession(session{Server: serverURL, Username: *username, Token: token}); err != nil {
		return err
	}

	fmt.Printf("logged in as %s\n", *username)
	return nil
}

func runMe(args []string) error {
	fs := flag.NewFlagSet("me", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	server := fs.String("server", "", "GitForge base URL")
	if err := fs.Parse(args); err != nil {
		return err
	}

	s, err := loadSession()
	if err != nil {
		return fmt.Errorf("load session: %w", err)
	}
	if *server != "" {
		s.Server = strings.TrimRight(*server, "/")
	}

	respBody, err := requestJSON(http.MethodGet, s.Server+"/api/auth/me", nil, s.Token)
	if err != nil {
		return err
	}

	os.Stdout.Write(respBody)
	if len(respBody) == 0 || respBody[len(respBody)-1] != '\n' {
		fmt.Println()
	}
	return nil
}

func runUserCreate(args []string) error {
	fs := flag.NewFlagSet("user-create", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	server := fs.String("server", "", "GitForge base URL")
	username := fs.String("username", "", "username")
	email := fs.String("email", "", "email")
	password := fs.String("password", "", "password")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *server == "" || *username == "" || *email == "" || *password == "" {
		return fmt.Errorf("server, username, email, and password are required")
	}

	serverURL := strings.TrimRight(*server, "/")
	respBody, err := requestJSON(http.MethodPost, serverURL+"/api/users", map[string]string{
		"username": *username,
		"email":    *email,
		"password": *password,
	}, "")
	if err != nil {
		return err
	}

	os.Stdout.Write(respBody)
	if len(respBody) == 0 || respBody[len(respBody)-1] != '\n' {
		fmt.Println()
	}
	return nil
}

func runRepoCreate(args []string) error {
	fs := flag.NewFlagSet("repo-create", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	server := fs.String("server", "", "GitForge base URL")
	owner := fs.String("owner", "", "owner username")
	name := fs.String("name", "", "repository name")
	description := fs.String("description", "", "repository description")
	visibility := fs.String("visibility", "private", "private or public")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *name == "" {
		return fmt.Errorf("name is required")
	}

	s, err := loadSession()
	if err != nil {
		return fmt.Errorf("load session: %w", err)
	}
	if *server != "" {
		s.Server = strings.TrimRight(*server, "/")
	}
	if *owner == "" {
		*owner = s.Username
	}

	respBody, err := requestJSON(http.MethodPost, s.Server+"/api/repos", map[string]string{
		"owner":       *owner,
		"name":        *name,
		"description": *description,
		"visibility":  *visibility,
	}, s.Token)
	if err != nil {
		return err
	}

	os.Stdout.Write(respBody)
	if len(respBody) == 0 || respBody[len(respBody)-1] != '\n' {
		fmt.Println()
	}
	return nil
}

func runGit(args []string) error {
	cmd := exec.Command("git", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}
		return err
	}
	return nil
}

func requestJSON(method, url string, payload interface{}, token string) ([]byte, error) {
	var bodyReader io.Reader
	if payload != nil {
		body, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("encode request body: %w", err)
		}
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if len(respBody) == 0 {
			return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	return respBody, nil
}
