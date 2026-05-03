package repositories

import (
	"context"
	"errors"
	"testing"

	"github.com/MattInnovates/gitforge/core/config"
	"github.com/MattInnovates/gitforge/core/domain"
	"github.com/MattInnovates/gitforge/core/storage/database"
)

func TestSQLiteUserStore_CreateAndGetByUsername(t *testing.T) {
	ctx := context.Background()
	cfg := config.Config{DatabaseDriver: "sqlite", DatabaseDSN: ":memory:"}
	db, err := database.OpenAndMigrate(ctx, cfg)
	if err != nil {
		t.Fatalf("open and migrate test db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	store := NewSQLiteUserStore(db)

	user := &domain.User{
		Username:     "matt",
		Email:        "matt@example.com",
		PasswordHash: "hash",
	}
	if err := store.Create(ctx, user); err != nil {
		t.Fatalf("create user: %v", err)
	}
	if user.ID == 0 {
		t.Fatal("expected non-zero user id")
	}

	got, err := store.GetByUsername(ctx, "matt")
	if err != nil {
		t.Fatalf("get by username: %v", err)
	}
	if got.Username != user.Username {
		t.Fatalf("expected username %q, got %q", user.Username, got.Username)
	}
}

func TestSQLiteRepositoryStore_CreateGetAndList(t *testing.T) {
	ctx := context.Background()
	cfg := config.Config{DatabaseDriver: "sqlite", DatabaseDSN: ":memory:"}
	db, err := database.OpenAndMigrate(ctx, cfg)
	if err != nil {
		t.Fatalf("open and migrate test db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	userStore := NewSQLiteUserStore(db)
	repoStore := NewSQLiteRepositoryStore(db)

	owner := &domain.User{
		Username:     "owner1",
		Email:        "owner1@example.com",
		PasswordHash: "hash",
	}
	if err := userStore.Create(ctx, owner); err != nil {
		t.Fatalf("create owner: %v", err)
	}

	repo := &domain.Repository{
		OwnerID:     owner.ID,
		Name:        "demo-repo",
		Description: "demo",
		Visibility:  domain.RepositoryVisibilityPublic,
	}
	if err := repoStore.Create(ctx, repo); err != nil {
		t.Fatalf("create repository: %v", err)
	}

	got, err := repoStore.GetByOwnerAndName(ctx, "owner1", "demo-repo")
	if err != nil {
		t.Fatalf("get repository by owner and name: %v", err)
	}
	if got.Name != repo.Name {
		t.Fatalf("expected repository name %q, got %q", repo.Name, got.Name)
	}
	if got.OwnerName != "owner1" {
		t.Fatalf("expected owner name %q, got %q", "owner1", got.OwnerName)
	}

	list, err := repoStore.ListByOwner(ctx, "owner1")
	if err != nil {
		t.Fatalf("list repositories by owner: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 repository, got %d", len(list))
	}
}

func TestSQLiteUserStore_GetByID_NotFound(t *testing.T) {
	ctx := context.Background()
	cfg := config.Config{DatabaseDriver: "sqlite", DatabaseDSN: ":memory:"}
	db, err := database.OpenAndMigrate(ctx, cfg)
	if err != nil {
		t.Fatalf("open and migrate test db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	store := NewSQLiteUserStore(db)
	_, err = store.GetByID(ctx, 999)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
