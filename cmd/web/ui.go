package main

import (
	"fmt"
	"html/template"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/MattInnovates/gitforge/core/auth"
	"github.com/MattInnovates/gitforge/core/domain"
	"github.com/MattInnovates/gitforge/core/storage/repositories"
)

type uiServer struct {
	users         repositories.UserStore
	repos         repositories.RepositoryStore
	auth          *auth.Service
	gitHTTPPort   int
	gitSSHPort    int
	publicTmpl    *template.Template
	profileTmpl   *template.Template
	repoTmpl      *template.Template
	adminTmpl     *template.Template
	webmasterTmpl *template.Template
}

type repoView struct {
	OwnerName   string
	Name        string
	Description string
	Visibility  domain.RepositoryVisibility
	DefaultRef  string
	ProfileURL  string
	RepoURL     string
	APIURL      string
	HTTPClone   string
	SSHClone    string
}

type homeData struct {
	OwnerFilter string
	Error       string
	Repos       []repoView
	AreaLabel   string
	BasePath    string
}

type profileData struct {
	Error       string
	OwnerFilter string
	User        domain.User
	Repos       []repoView
	RepoCount   int
}

type repoPageData struct {
	Error      string
	Repo       repoView
	OtherRepos []repoView
}

func NewUIHandler(users repositories.UserStore, repos repositories.RepositoryStore, authService *auth.Service, gitHTTPPort int, gitSSHPort int) http.Handler {
	publicTemplatePath, err := resolveWebPath(
		filepath.Join("cmd", "web", "templates", "public.html"),
		filepath.Join("templates", "public.html"),
	)
	if err != nil {
		panic(err)
	}

	adminTemplatePath, err := resolveWebPath(
		filepath.Join("cmd", "web", "templates", "admin.html"),
		filepath.Join("templates", "admin.html"),
	)
	if err != nil {
		panic(err)
	}

	profileTemplatePath, err := resolveWebPath(
		filepath.Join("cmd", "web", "templates", "profile.html"),
		filepath.Join("templates", "profile.html"),
	)
	if err != nil {
		panic(err)
	}

	repoTemplatePath, err := resolveWebPath(
		filepath.Join("cmd", "web", "templates", "repo.html"),
		filepath.Join("templates", "repo.html"),
	)
	if err != nil {
		panic(err)
	}

	staticDirPath, err := resolveWebPath(
		filepath.Join("cmd", "web", "static"),
		"static",
	)
	if err != nil {
		panic(err)
	}

	s := &uiServer{
		users:         users,
		repos:         repos,
		auth:          authService,
		gitHTTPPort:   gitHTTPPort,
		gitSSHPort:    gitSSHPort,
		publicTmpl:    template.Must(template.ParseFiles(publicTemplatePath)),
		profileTmpl:   template.Must(template.ParseFiles(profileTemplatePath)),
		repoTmpl:      template.Must(template.ParseFiles(repoTemplatePath)),
		adminTmpl:     template.Must(template.ParseFiles(adminTemplatePath)),
		webmasterTmpl: template.Must(template.ParseFiles(adminTemplatePath)),
	}

	mux := http.NewServeMux()
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(staticDirPath))))
	mux.HandleFunc("/", s.handlePublicHome)
	mux.HandleFunc("/u/", s.handleUserRoutes)
	mux.HandleFunc("/admin", s.handleAdminHome)
	mux.HandleFunc("/webmaster", s.handleWebmasterHome)
	mux.HandleFunc("/admin/users", s.handleAdminCreateUser)
	mux.HandleFunc("/admin/repos", s.handleAdminCreateRepo)
	mux.HandleFunc("/webmaster/users", s.handleWebmasterCreateUser)
	mux.HandleFunc("/webmaster/repos", s.handleWebmasterCreateRepo)
	return mux
}

func resolveWebPath(candidates ...string) (string, error) {
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}

		if _, err := os.Stat(candidate); err == nil {
			abs, absErr := filepath.Abs(candidate)
			if absErr != nil {
				return "", fmt.Errorf("resolve path %s: %w", candidate, absErr)
			}
			return abs, nil
		}
	}

	return "", fmt.Errorf("unable to resolve web path from candidates: %v", candidates)
}

func (s *uiServer) handlePublicHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	owner := strings.TrimSpace(r.URL.Query().Get("owner"))
	errCode := strings.TrimSpace(r.URL.Query().Get("err"))
	data := homeData{OwnerFilter: owner, Error: uiErrorMessage(errCode), AreaLabel: "Public", BasePath: "/"}

	if owner != "" {
		repos, err := s.repos.ListByOwner(r.Context(), owner)
		if err != nil {
			data.Error = "failed to load repositories"
		} else {
			data.Repos = s.buildRepoViews(r, repos)
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = s.publicTmpl.Execute(w, data)
}

func (s *uiServer) handleUserRoutes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/u/")
	path = strings.Trim(path, "/")
	if path == "" {
		http.NotFound(w, r)
		return
	}

	parts := strings.Split(path, "/")
	if len(parts) == 1 {
		s.handleProfileView(w, r, parts[0])
		return
	}
	if len(parts) == 2 {
		s.handleRepoView(w, r, parts[0], parts[1])
		return
	}

	http.NotFound(w, r)
}

func (s *uiServer) handleProfileView(w http.ResponseWriter, r *http.Request, ownerRaw string) {
	owner, err := url.PathUnescape(ownerRaw)
	if err != nil || strings.TrimSpace(owner) == "" {
		http.NotFound(w, r)
		return
	}

	user, err := s.users.GetByUsername(r.Context(), owner)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	repos, err := s.repos.ListByOwner(r.Context(), owner)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = s.profileTmpl.Execute(w, profileData{Error: "Failed to load repositories", User: user})
		return
	}

	views := s.buildRepoViews(r, repos)
	data := profileData{
		OwnerFilter: owner,
		User:        user,
		Repos:       views,
		RepoCount:   len(views),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = s.profileTmpl.Execute(w, data)
}

func (s *uiServer) handleRepoView(w http.ResponseWriter, r *http.Request, ownerRaw string, repoRaw string) {
	owner, err := url.PathUnescape(ownerRaw)
	if err != nil || strings.TrimSpace(owner) == "" {
		http.NotFound(w, r)
		return
	}

	repoName, err := url.PathUnescape(repoRaw)
	if err != nil || strings.TrimSpace(repoName) == "" {
		http.NotFound(w, r)
		return
	}

	repo, err := s.repos.GetByOwnerAndName(r.Context(), owner, repoName)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	view := s.buildRepoViews(r, []domain.Repository{repo})[0]

	ownerRepos, err := s.repos.ListByOwner(r.Context(), owner)
	if err != nil {
		ownerRepos = []domain.Repository{}
	}

	other := make([]domain.Repository, 0)
	for _, item := range ownerRepos {
		if item.Name == repo.Name {
			continue
		}
		other = append(other, item)
	}

	data := repoPageData{
		Repo:       view,
		OtherRepos: s.buildRepoViews(r, other),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = s.repoTmpl.Execute(w, data)
}

func (s *uiServer) handleAdminHome(w http.ResponseWriter, r *http.Request) {
	s.renderBackendHome(w, r, "/admin", "Admin", s.adminTmpl)
}

func (s *uiServer) handleWebmasterHome(w http.ResponseWriter, r *http.Request) {
	s.renderBackendHome(w, r, "/webmaster", "Webmaster", s.webmasterTmpl)
}

func (s *uiServer) renderBackendHome(w http.ResponseWriter, r *http.Request, basePath string, areaLabel string, tmpl *template.Template) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	owner := strings.TrimSpace(r.URL.Query().Get("owner"))
	errCode := strings.TrimSpace(r.URL.Query().Get("err"))
	data := homeData{OwnerFilter: owner, Error: uiErrorMessage(errCode), AreaLabel: areaLabel, BasePath: basePath}

	if owner != "" {
		repos, err := s.repos.ListByOwner(r.Context(), owner)
		if err != nil {
			data.Error = "failed to load repositories"
		} else {
			data.Repos = s.buildRepoViews(r, repos)
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.Execute(w, data)
}

func uiErrorMessage(code string) string {
	switch code {
	case "invalid_form":
		return "The submitted form is invalid. Please try again."
	case "missing_user_fields":
		return "Username, email, and password are required."
	case "secure_password_failed":
		return "Unable to secure the password. Please retry."
	case "create_user_failed":
		return "Failed to create user. The username/email may already exist."
	case "missing_repo_fields":
		return "Owner and repository name are required."
	case "owner_not_found":
		return "Owner not found. Create that user first."
	case "create_repo_failed":
		return "Failed to create repository. It may already exist for this owner."
	default:
		return ""
	}
}

func (s *uiServer) handleAdminCreateUser(w http.ResponseWriter, r *http.Request) {
	s.handleCreateUser(w, r, "/admin")
}

func (s *uiServer) handleWebmasterCreateUser(w http.ResponseWriter, r *http.Request) {
	s.handleCreateUser(w, r, "/webmaster")
}

func (s *uiServer) handleCreateUser(w http.ResponseWriter, r *http.Request, basePath string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, basePath+"?err=invalid_form", http.StatusSeeOther)
		return
	}

	username := strings.TrimSpace(r.FormValue("username"))
	email := strings.TrimSpace(r.FormValue("email"))
	password := strings.TrimSpace(r.FormValue("password"))
	if username == "" || email == "" || password == "" {
		http.Redirect(w, r, basePath+"?err=missing_user_fields", http.StatusSeeOther)
		return
	}

	user := &domain.User{
		Username:     username,
		Email:        email,
		PasswordHash: password,
	}

	hash, err := s.auth.HashPassword(password)
	if err != nil {
		http.Redirect(w, r, basePath+"?err=secure_password_failed", http.StatusSeeOther)
		return
	}
	user.PasswordHash = hash

	if err := s.users.Create(r.Context(), user); err != nil {
		http.Redirect(w, r, basePath+"?err=create_user_failed", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, basePath+"?owner="+username, http.StatusSeeOther)
}

func (s *uiServer) handleAdminCreateRepo(w http.ResponseWriter, r *http.Request) {
	s.handleCreateRepo(w, r, "/admin")
}

func (s *uiServer) handleWebmasterCreateRepo(w http.ResponseWriter, r *http.Request) {
	s.handleCreateRepo(w, r, "/webmaster")
}

func (s *uiServer) handleCreateRepo(w http.ResponseWriter, r *http.Request, basePath string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, basePath+"?err=invalid_form", http.StatusSeeOther)
		return
	}

	ownerName := strings.TrimSpace(r.FormValue("owner"))
	repoName := strings.TrimSpace(r.FormValue("name"))
	description := strings.TrimSpace(r.FormValue("description"))
	visibility := domain.RepositoryVisibility(strings.TrimSpace(r.FormValue("visibility")))
	if ownerName == "" || repoName == "" {
		http.Redirect(w, r, basePath+"?err=missing_repo_fields", http.StatusSeeOther)
		return
	}
	if visibility == "" {
		visibility = domain.RepositoryVisibilityPrivate
	}

	owner, err := s.users.GetByUsername(r.Context(), ownerName)
	if err != nil {
		http.Redirect(w, r, basePath+"?err=owner_not_found", http.StatusSeeOther)
		return
	}

	repo := &domain.Repository{
		OwnerID:     owner.ID,
		Name:        repoName,
		Description: description,
		Visibility:  visibility,
	}
	if err := s.repos.Create(r.Context(), repo); err != nil {
		http.Redirect(w, r, basePath+"?owner="+ownerName+"&err=create_repo_failed", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, basePath+"?owner="+ownerName, http.StatusSeeOther)
}

func (s *uiServer) buildRepoViews(r *http.Request, repos []domain.Repository) []repoView {
	host := hostWithoutPort(r.Host)
	items := make([]repoView, 0, len(repos))

	for _, repo := range repos {
		ownerEscaped := url.PathEscape(repo.OwnerName)
		nameEscaped := url.PathEscape(repo.Name)
		apiURL := "/api/repos/" + ownerEscaped + "/" + nameEscaped

		items = append(items, repoView{
			OwnerName:   repo.OwnerName,
			Name:        repo.Name,
			Description: repo.Description,
			Visibility:  repo.Visibility,
			DefaultRef:  repo.DefaultRef,
			ProfileURL:  "/u/" + ownerEscaped,
			RepoURL:     "/u/" + ownerEscaped + "/" + nameEscaped,
			APIURL:      apiURL,
			HTTPClone:   fmt.Sprintf("http://%s:%d/%s/%s.git", host, s.gitHTTPPort, ownerEscaped, nameEscaped),
			SSHClone:    fmt.Sprintf("ssh://git@%s:%d/%s/%s.git", host, s.gitSSHPort, ownerEscaped, nameEscaped),
		})
	}

	return items
}

func hostWithoutPort(host string) string {
	if parsedHost, _, err := net.SplitHostPort(host); err == nil {
		return strings.Trim(parsedHost, "[]")
	}

	return strings.Trim(host, "[]")
}
