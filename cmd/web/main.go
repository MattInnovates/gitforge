package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/MattInnovates/gitforge/api"
	"github.com/MattInnovates/gitforge/core/auth"
	"github.com/MattInnovates/gitforge/core/config"
	"github.com/MattInnovates/gitforge/core/storage/database"
	"github.com/MattInnovates/gitforge/core/storage/repositories"
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

	userStore := repositories.NewSQLiteUserStore(db)
	repoStore := repositories.NewSQLiteRepositoryStore(db)
	authService := auth.NewService(cfg.AuthTokenSecret, time.Duration(cfg.AuthTokenTTLMinutes)*time.Minute)
	apiHandler := api.SetupRoutes(userStore, repoStore, authService)
	uiHandler := NewUIHandler(userStore, repoStore, authService, cfg.GitHTTPPort, cfg.GitSSHPort, cfg.ReposPath)

	router := http.NewServeMux()
	router.Handle("/api/", apiHandler)
	router.Handle("/api", apiHandler)
	router.Handle("/", uiHandler)

	addr := fmt.Sprintf(":%d", cfg.WebPort)
	log.Printf("web server listening on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("web server stopped: %v", err)
	}
}
