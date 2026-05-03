package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/MattInnovates/gitforge/core/auth"
	"github.com/MattInnovates/gitforge/core/domain"
	"github.com/MattInnovates/gitforge/core/storage/repositories"
)

type server struct {
	users repositories.UserStore
	repos repositories.RepositoryStore
	auth  *auth.Service
}

func SetupRoutes(users repositories.UserStore, repos repositories.RepositoryStore, authService *auth.Service) http.Handler {
	s := &server{users: users, repos: repos, auth: authService}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/healthz", s.handleHealthz)
	mux.HandleFunc("/api/auth/login", s.handleAuthLogin)
	mux.HandleFunc("/api/auth/me", s.handleAuthMe)
	mux.HandleFunc("/api/users", s.handleUsers)
	mux.HandleFunc("/api/users/", s.handleUserByUsername)
	mux.HandleFunc("/api/repos", s.handleRepos)
	mux.HandleFunc("/api/repos/", s.handleRepoByOwnerAndName)

	return mux
}

func (s *server) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *server) handleAuthLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := decodeJSONBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	username := strings.TrimSpace(req.Username)
	if username == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "username and password are required")
		return
	}

	user, err := s.users.GetByUsername(r.Context(), username)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if !s.auth.VerifyPassword(req.Password, user.PasswordHash) {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := s.auth.IssueToken(user.ID, user.Username, time.Now().UTC())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to issue token")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (s *server) handleAuthMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}

	claims, ok := s.authenticateRequest(w, r)
	if !ok {
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user_id":  claims.UserID,
		"username": claims.Username,
		"exp":      claims.ExpiresAt,
	})
}

func (s *server) handleUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}

	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := decodeJSONBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)
	if req.Username == "" || req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "username, email, and password are required")
		return
	}

	user := &domain.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: req.Password,
	}

	passwordHash, err := s.auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to secure password")
		return
	}
	user.PasswordHash = passwordHash

	if err := s.users.Create(r.Context(), user); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	})
}

func (s *server) handleUserByUsername(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}

	username := strings.TrimPrefix(r.URL.Path, "/api/users/")
	username = strings.TrimSpace(username)
	if username == "" {
		writeError(w, http.StatusBadRequest, "username is required")
		return
	}

	user, err := s.users.GetByUsername(r.Context(), username)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to fetch user")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	})
}

func (s *server) handleRepos(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.handleCreateRepo(w, r)
	case http.MethodGet:
		s.handleListReposByOwner(w, r)
	default:
		writeMethodNotAllowed(w)
	}
}

func (s *server) handleCreateRepo(w http.ResponseWriter, r *http.Request) {
	claims, ok := s.authenticateRequest(w, r)
	if !ok {
		return
	}

	var req struct {
		Owner       string `json:"owner"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Visibility  string `json:"visibility"`
	}
	if err := decodeJSONBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	req.Owner = strings.TrimSpace(req.Owner)
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.Owner == "" {
		req.Owner = claims.Username
	}
	if req.Owner != claims.Username {
		writeError(w, http.StatusForbidden, "owner must match authenticated user")
		return
	}

	visibility := domain.RepositoryVisibility(strings.TrimSpace(req.Visibility))
	if visibility == "" {
		visibility = domain.RepositoryVisibilityPrivate
	}
	if visibility != domain.RepositoryVisibilityPrivate && visibility != domain.RepositoryVisibilityPublic {
		writeError(w, http.StatusBadRequest, "visibility must be private or public")
		return
	}

	owner, err := s.users.GetByUsername(r.Context(), req.Owner)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			writeError(w, http.StatusNotFound, "owner not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to fetch owner")
		return
	}

	repo := &domain.Repository{
		OwnerID:     owner.ID,
		Name:        req.Name,
		Description: req.Description,
		Visibility:  visibility,
	}
	if err := s.repos.Create(r.Context(), repo); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create repository")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":          repo.ID,
		"owner":       owner.Username,
		"name":        repo.Name,
		"description": repo.Description,
		"visibility":  repo.Visibility,
		"default_ref": repo.DefaultRef,
		"created_at":  repo.CreatedAt,
		"updated_at":  repo.UpdatedAt,
	})
}

func (s *server) handleListReposByOwner(w http.ResponseWriter, r *http.Request) {
	owner := strings.TrimSpace(r.URL.Query().Get("owner"))
	if owner == "" {
		writeError(w, http.StatusBadRequest, "owner query parameter is required")
		return
	}

	repos, err := s.repos.ListByOwner(r.Context(), owner)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list repositories")
		return
	}

	out := make([]map[string]interface{}, 0, len(repos))
	for _, repo := range repos {
		out = append(out, map[string]interface{}{
			"id":          repo.ID,
			"owner":       repo.OwnerName,
			"name":        repo.Name,
			"description": repo.Description,
			"visibility":  repo.Visibility,
			"default_ref": repo.DefaultRef,
			"created_at":  repo.CreatedAt,
			"updated_at":  repo.UpdatedAt,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"repositories": out})
}

func (s *server) handleRepoByOwnerAndName(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/repos/"), "/")
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		writeError(w, http.StatusBadRequest, "path must be /api/repos/{owner}/{name}")
		return
	}

	repo, err := s.repos.GetByOwnerAndName(r.Context(), parts[0], parts[1])
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			writeError(w, http.StatusNotFound, "repository not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to fetch repository")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":          repo.ID,
		"owner":       repo.OwnerName,
		"name":        repo.Name,
		"description": repo.Description,
		"visibility":  repo.Visibility,
		"default_ref": repo.DefaultRef,
		"created_at":  repo.CreatedAt,
		"updated_at":  repo.UpdatedAt,
	})
}

func decodeJSONBody(r *http.Request, dst interface{}) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return err
	}
	return nil
}

func writeMethodNotAllowed(w http.ResponseWriter) {
	writeError(w, http.StatusMethodNotAllowed, "method not allowed")
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, statusCode int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(v)
}

func (s *server) authenticateRequest(w http.ResponseWriter, r *http.Request) (auth.Claims, bool) {
	header := strings.TrimSpace(r.Header.Get("Authorization"))
	if !strings.HasPrefix(strings.ToLower(header), "bearer ") {
		writeError(w, http.StatusUnauthorized, "missing bearer token")
		return auth.Claims{}, false
	}

	token := strings.TrimSpace(header[len("Bearer "):])
	if token == "" {
		writeError(w, http.StatusUnauthorized, "missing bearer token")
		return auth.Claims{}, false
	}

	claims, err := s.auth.ParseToken(token, time.Now().UTC())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid token")
		return auth.Claims{}, false
	}

	return claims, true
}
