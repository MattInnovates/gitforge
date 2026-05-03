package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/MattInnovates/gitforge/core/auth"
	"github.com/MattInnovates/gitforge/core/config"
	"github.com/MattInnovates/gitforge/core/storage/database"
	"github.com/MattInnovates/gitforge/core/storage/repositories"
)

func newTestHandler(t *testing.T) http.Handler {
	t.Helper()

	cfg := config.Config{DatabaseDriver: "sqlite", DatabaseDSN: ":memory:"}
	db, err := database.OpenAndMigrate(context.Background(), cfg)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	userStore := repositories.NewSQLiteUserStore(db)
	repoStore := repositories.NewSQLiteRepositoryStore(db)
	authService := auth.NewService("test-secret", time.Hour)
	return SetupRoutes(userStore, repoStore, authService)
}

func createUser(t *testing.T, h http.Handler, username string) {
	t.Helper()

	createReq := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(`{"username":"`+username+`","email":"`+username+`@example.com","password":"secret"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createRec := httptest.NewRecorder()
	h.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusCreated, createRec.Code, createRec.Body.String())
	}
}

func loginToken(t *testing.T, h http.Handler, username string) string {
	t.Helper()

	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"username":"`+username+`","password":"secret"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	h.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, loginRec.Code, loginRec.Body.String())
	}

	var body map[string]string
	if err := json.Unmarshal(loginRec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	token := body["token"]
	if token == "" {
		t.Fatal("expected token in login response")
	}
	return token
}

func TestCreateAndGetUser(t *testing.T) {
	h := newTestHandler(t)

	createUser(t, h, "matt")

	getReq := httptest.NewRequest(http.MethodGet, "/api/users/matt", nil)
	getRec := httptest.NewRecorder()
	h.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, getRec.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(getRec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["username"] != "matt" {
		t.Fatalf("expected username matt, got %v", body["username"])
	}
}

func TestCreateRepoGetAndListByOwner(t *testing.T) {
	h := newTestHandler(t)

	createUser(t, h, "owner1")
	token := loginToken(t, h, "owner1")

	createRepo := httptest.NewRequest(http.MethodPost, "/api/repos", strings.NewReader(`{"owner":"owner1","name":"demo-repo","description":"demo","visibility":"public"}`))
	createRepo.Header.Set("Content-Type", "application/json")
	createRepo.Header.Set("Authorization", "Bearer "+token)
	repoRec := httptest.NewRecorder()
	h.ServeHTTP(repoRec, createRepo)
	if repoRec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusCreated, repoRec.Code, repoRec.Body.String())
	}

	getRepo := httptest.NewRequest(http.MethodGet, "/api/repos/owner1/demo-repo", nil)
	getRepoRec := httptest.NewRecorder()
	h.ServeHTTP(getRepoRec, getRepo)
	if getRepoRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, getRepoRec.Code)
	}

	listRepos := httptest.NewRequest(http.MethodGet, "/api/repos?owner=owner1", nil)
	listRec := httptest.NewRecorder()
	h.ServeHTTP(listRec, listRepos)
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, listRec.Code)
	}

	var body map[string][]map[string]interface{}
	if err := json.Unmarshal(listRec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(body["repositories"]) != 1 {
		t.Fatalf("expected 1 repository, got %d", len(body["repositories"]))
	}
}

func TestGetMissingRepoReturnsNotFound(t *testing.T) {
	h := newTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/repos/owner1/missing", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestAuthLoginAndMe(t *testing.T) {
	h := newTestHandler(t)
	createUser(t, h, "alice")
	token := loginToken(t, h, "alice")

	meReq := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+token)
	meRec := httptest.NewRecorder()
	h.ServeHTTP(meRec, meReq)

	if meRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, meRec.Code, meRec.Body.String())
	}

	var body map[string]interface{}
	if err := json.Unmarshal(meRec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode me response: %v", err)
	}
	if body["username"] != "alice" {
		t.Fatalf("expected username alice, got %v", body["username"])
	}
}
