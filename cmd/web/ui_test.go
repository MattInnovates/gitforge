package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/MattInnovates/gitforge/core/auth"
	"github.com/MattInnovates/gitforge/core/config"
	"github.com/MattInnovates/gitforge/core/storage/database"
	"github.com/MattInnovates/gitforge/core/storage/repositories"
)

func newWebTestHandler(t *testing.T) http.Handler {
	t.Helper()

	cfg := config.Config{DatabaseDriver: "sqlite", DatabaseDSN: ":memory:"}
	db, err := database.OpenAndMigrate(context.Background(), cfg)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	users := repositories.NewSQLiteUserStore(db)
	repos := repositories.NewSQLiteRepositoryStore(db)
	authService := auth.NewService("test-secret", time.Hour)
	return NewUIHandler(users, repos, authService, 8080, 2222)
}

func TestHomePageRenders(t *testing.T) {
	h := newWebTestHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d got %d", http.StatusOK, rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "GitForge") {
		t.Fatalf("expected page body to contain title")
	}
}

func TestAdminPageRenders(t *testing.T) {
	h := newWebTestHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d got %d", http.StatusOK, rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Admin Workspace") {
		t.Fatalf("expected admin page title")
	}
}

func TestCreateUserRedirects(t *testing.T) {
	h := newWebTestHandler(t)
	form := url.Values{}
	form.Set("username", "matt")
	form.Set("email", "matt@example.com")
	form.Set("password", "secret")

	req := httptest.NewRequest(http.MethodPost, "/admin/users", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected status %d got %d", http.StatusSeeOther, rec.Code)
	}
	if location := rec.Header().Get("Location"); location != "/admin?owner=matt" {
		t.Fatalf("unexpected redirect location %q", location)
	}
}

func TestCreateRepoRedirects(t *testing.T) {
	h := newWebTestHandler(t)

	createUserForm := url.Values{}
	createUserForm.Set("username", "owner1")
	createUserForm.Set("email", "owner1@example.com")
	createUserForm.Set("password", "secret")
	createUserReq := httptest.NewRequest(http.MethodPost, "/admin/users", strings.NewReader(createUserForm.Encode()))
	createUserReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	createUserRec := httptest.NewRecorder()
	h.ServeHTTP(createUserRec, createUserReq)

	if createUserRec.Code != http.StatusSeeOther {
		t.Fatalf("expected user create status %d got %d", http.StatusSeeOther, createUserRec.Code)
	}

	createRepoForm := url.Values{}
	createRepoForm.Set("owner", "owner1")
	createRepoForm.Set("name", "demo")
	createRepoForm.Set("description", "test")
	createRepoForm.Set("visibility", "public")
	createRepoReq := httptest.NewRequest(http.MethodPost, "/admin/repos", strings.NewReader(createRepoForm.Encode()))
	createRepoReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	createRepoRec := httptest.NewRecorder()
	h.ServeHTTP(createRepoRec, createRepoReq)

	if createRepoRec.Code != http.StatusSeeOther {
		t.Fatalf("expected repo create status %d got %d", http.StatusSeeOther, createRepoRec.Code)
	}
	if location := createRepoRec.Header().Get("Location"); location != "/admin?owner=owner1" {
		t.Fatalf("unexpected redirect location %q", location)
	}
}

func TestProfilePageRenders(t *testing.T) {
	h := newWebTestHandler(t)
	seedUserAndRepo(t, h, "owner2", "owner2@example.com", "sample")

	req := httptest.NewRequest(http.MethodGet, "/u/owner2", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d got %d", http.StatusOK, rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "owner2") {
		t.Fatalf("expected profile page to contain username")
	}
}

func TestRepoPageRenders(t *testing.T) {
	h := newWebTestHandler(t)
	seedUserAndRepo(t, h, "owner3", "owner3@example.com", "project")

	req := httptest.NewRequest(http.MethodGet, "/u/owner3/project", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d got %d", http.StatusOK, rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "owner3") || !strings.Contains(body, "project") {
		t.Fatalf("expected repo page to contain owner and repository name")
	}
}

func TestProfileAndRepoNotFound(t *testing.T) {
	h := newWebTestHandler(t)

	profileReq := httptest.NewRequest(http.MethodGet, "/u/missing-owner", nil)
	profileRec := httptest.NewRecorder()
	h.ServeHTTP(profileRec, profileReq)
	if profileRec.Code != http.StatusNotFound {
		t.Fatalf("expected profile status %d got %d", http.StatusNotFound, profileRec.Code)
	}

	repoReq := httptest.NewRequest(http.MethodGet, "/u/missing-owner/missing-repo", nil)
	repoRec := httptest.NewRecorder()
	h.ServeHTTP(repoRec, repoReq)
	if repoRec.Code != http.StatusNotFound {
		t.Fatalf("expected repo status %d got %d", http.StatusNotFound, repoRec.Code)
	}
}

func seedUserAndRepo(t *testing.T, h http.Handler, username string, email string, repoName string) {
	t.Helper()

	createUserForm := url.Values{}
	createUserForm.Set("username", username)
	createUserForm.Set("email", email)
	createUserForm.Set("password", "secret")
	createUserReq := httptest.NewRequest(http.MethodPost, "/admin/users", strings.NewReader(createUserForm.Encode()))
	createUserReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	createUserRec := httptest.NewRecorder()
	h.ServeHTTP(createUserRec, createUserReq)
	if createUserRec.Code != http.StatusSeeOther {
		t.Fatalf("expected user create status %d got %d", http.StatusSeeOther, createUserRec.Code)
	}

	createRepoForm := url.Values{}
	createRepoForm.Set("owner", username)
	createRepoForm.Set("name", repoName)
	createRepoForm.Set("description", "repo description")
	createRepoForm.Set("visibility", "public")
	createRepoReq := httptest.NewRequest(http.MethodPost, "/admin/repos", strings.NewReader(createRepoForm.Encode()))
	createRepoReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	createRepoRec := httptest.NewRecorder()
	h.ServeHTTP(createRepoRec, createRepoReq)
	if createRepoRec.Code != http.StatusSeeOther {
		t.Fatalf("expected repo create status %d got %d", http.StatusSeeOther, createRepoRec.Code)
	}
}
