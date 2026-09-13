package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/khanhvunguyen/vendor-onboarding-tracker/internal/database"
)

type Repository struct {
	database database.Queryer
}

func NewRepository(database database.Queryer) *Repository {
	return &Repository{database: database}
}

func (r *Repository) FindByEmail(ctx context.Context, email string) (coordinatorWithPassword, error) {
	const query = `
		SELECT id, name, email, password_hash
		FROM coordinators
		WHERE email = $1
	`

	var coordinator coordinatorWithPassword
	err := r.database.QueryRow(ctx, query, email).Scan(
		&coordinator.ID,
		&coordinator.Name,
		&coordinator.Email,
		&coordinator.PasswordHash,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return coordinatorWithPassword{}, errCoordinatorNotFound
	}
	if err != nil {
		return coordinatorWithPassword{}, fmt.Errorf("find coordinator by email: %w", err)
	}

	return coordinator, nil
}
