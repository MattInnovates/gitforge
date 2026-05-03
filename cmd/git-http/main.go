package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/cgi"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/MattInnovates/gitforge/core/config"
	"github.com/MattInnovates/gitforge/core/storage/database"
)

func main() {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := database.OpenAndMigrate(context.Background(), cfg)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	repoRoot, err := filepath.Abs(cfg.ReposPath)
	if err != nil {
		log.Fatalf("resolve repos path: %v", err)
	}

	if err := os.MkdirAll(repoRoot, 0o755); err != nil {
		log.Fatalf("create repos path: %v", err)
	}

	gitPath, err := exec.LookPath("git")
	if err != nil {
		log.Fatalf("git executable not found in PATH: %v", err)
	}

	cgiHandler := &cgi.Handler{
		Path:       gitPath,
		Args:       []string{"http-backend"},
		Env:        []string{"GIT_PROJECT_ROOT=" + repoRoot, "GIT_HTTP_EXPORT_ALL=1"},
		InheritEnv: []string{"PATH", "SYSTEMROOT", "WINDIR", "HOME", "USERPROFILE", "TMP", "TEMP"},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.Handle("/", requestLogger(cgiHandler))

	addr := fmt.Sprintf(":%d", cfg.GitHTTPPort)
	log.Printf("git-http listening on %s (repos=%s)", addr, repoRoot)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("git-http server stopped: %v", err)
	}
}

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
