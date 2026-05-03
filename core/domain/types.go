package domain

import "time"

// User represents an account that can own and access repositories.
type User struct {
	ID           int64
	Username     string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// RepositoryVisibility controls who can read a repository.
type RepositoryVisibility string

const (
	RepositoryVisibilityPrivate RepositoryVisibility = "private"
	RepositoryVisibilityPublic  RepositoryVisibility = "public"
)

// Repository represents a Git repository tracked by GitForge.
type Repository struct {
	ID          int64
	OwnerID     int64
	OwnerName   string
	Name        string
	Description string
	Visibility  RepositoryVisibility
	DefaultRef  string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
