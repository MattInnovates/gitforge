package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/MattInnovates/gitforge/core/domain"
)

var ErrNotFound = errors.New("entity not found")

type UserStore interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id int64) (domain.User, error)
	GetByUsername(ctx context.Context, username string) (domain.User, error)
}

type RepositoryStore interface {
	Create(ctx context.Context, repo *domain.Repository) error
	GetByOwnerAndName(ctx context.Context, ownerName, repoName string) (domain.Repository, error)
	ListByOwner(ctx context.Context, ownerName string) ([]domain.Repository, error)
}

type SQLiteUserStore struct {
	db *sql.DB
}

type SQLiteRepositoryStore struct {
	db *sql.DB
}

func NewSQLiteUserStore(db *sql.DB) *SQLiteUserStore {
	return &SQLiteUserStore{db: db}
}

func NewSQLiteRepositoryStore(db *sql.DB) *SQLiteRepositoryStore {
	return &SQLiteRepositoryStore{db: db}
}

func (s *SQLiteUserStore) Create(ctx context.Context, user *domain.User) error {
	now := time.Now().UTC()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = now
	}

	result, err := s.db.ExecContext(ctx,
		`INSERT INTO users (username, email, password_hash, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?)`,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get user id: %w", err)
	}
	user.ID = id

	return nil
}

func (s *SQLiteUserStore) GetByID(ctx context.Context, id int64) (domain.User, error) {
	var user domain.User
	err := s.db.QueryRowContext(ctx,
		`SELECT id, username, email, password_hash, created_at, updated_at
		 FROM users
		 WHERE id = ?`,
		id,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, ErrNotFound
		}
		return domain.User{}, fmt.Errorf("get user by id: %w", err)
	}

	return user, nil
}

func (s *SQLiteUserStore) GetByUsername(ctx context.Context, username string) (domain.User, error) {
	var user domain.User
	err := s.db.QueryRowContext(ctx,
		`SELECT id, username, email, password_hash, created_at, updated_at
		 FROM users
		 WHERE username = ?`,
		username,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, ErrNotFound
		}
		return domain.User{}, fmt.Errorf("get user by username: %w", err)
	}

	return user, nil
}

func (s *SQLiteRepositoryStore) Create(ctx context.Context, repo *domain.Repository) error {
	now := time.Now().UTC()
	if repo.Visibility == "" {
		repo.Visibility = domain.RepositoryVisibilityPrivate
	}
	if repo.DefaultRef == "" {
		repo.DefaultRef = "refs/heads/main"
	}
	if repo.CreatedAt.IsZero() {
		repo.CreatedAt = now
	}
	if repo.UpdatedAt.IsZero() {
		repo.UpdatedAt = now
	}

	result, err := s.db.ExecContext(ctx,
		`INSERT INTO repositories (owner_id, name, description, visibility, default_ref, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		repo.OwnerID,
		repo.Name,
		repo.Description,
		repo.Visibility,
		repo.DefaultRef,
		repo.CreatedAt,
		repo.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert repository: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get repository id: %w", err)
	}
	repo.ID = id

	return nil
}

func (s *SQLiteRepositoryStore) GetByOwnerAndName(ctx context.Context, ownerName, repoName string) (domain.Repository, error) {
	var repo domain.Repository
	err := s.db.QueryRowContext(ctx,
		`SELECT r.id, r.owner_id, u.username, r.name, r.description, r.visibility, r.default_ref, r.created_at, r.updated_at
		 FROM repositories r
		 INNER JOIN users u ON u.id = r.owner_id
		 WHERE u.username = ? AND r.name = ?`,
		ownerName,
		repoName,
	).Scan(
		&repo.ID,
		&repo.OwnerID,
		&repo.OwnerName,
		&repo.Name,
		&repo.Description,
		&repo.Visibility,
		&repo.DefaultRef,
		&repo.CreatedAt,
		&repo.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Repository{}, ErrNotFound
		}
		return domain.Repository{}, fmt.Errorf("get repository by owner and name: %w", err)
	}

	return repo, nil
}

func (s *SQLiteRepositoryStore) ListByOwner(ctx context.Context, ownerName string) ([]domain.Repository, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT r.id, r.owner_id, u.username, r.name, r.description, r.visibility, r.default_ref, r.created_at, r.updated_at
		 FROM repositories r
		 INNER JOIN users u ON u.id = r.owner_id
		 WHERE u.username = ?
		 ORDER BY r.name ASC`,
		ownerName,
	)
	if err != nil {
		return nil, fmt.Errorf("list repositories by owner: %w", err)
	}
	defer rows.Close()

	repos := make([]domain.Repository, 0)
	for rows.Next() {
		var repo domain.Repository
		if err := rows.Scan(
			&repo.ID,
			&repo.OwnerID,
			&repo.OwnerName,
			&repo.Name,
			&repo.Description,
			&repo.Visibility,
			&repo.DefaultRef,
			&repo.CreatedAt,
			&repo.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan repository: %w", err)
		}
		repos = append(repos, repo)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate repositories: %w", err)
	}

	return repos, nil
}
