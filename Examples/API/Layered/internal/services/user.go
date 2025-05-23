package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"example.com/examples/api/layered/internal/models"
	"github.com/jmoiron/sqlx"
)

// UsersService is a service capable of performing CRUD operations for
// models.User models.
type UsersService struct {
	logger *slog.Logger
	db     *sqlx.DB
}

// NewUsersService creates a new UsersService and returns a pointer to it.
func NewUsersService(logger *slog.Logger, db *sqlx.DB) *UsersService {
	return &UsersService{
		logger: logger,
		db:     db,
	}
}

// CreateUser attempts to create the provided user, returning a fully hydrated
// models.User or an error.
func (s *UsersService) CreateUser(ctx context.Context, user models.User) (models.User, error) {
	s.logger.DebugContext(ctx, "Creating user", "name", user.Name)

	row := s.db.QueryRowContext(
		ctx,
		`
		INSERT INTO users (name, email, password) VALUES ($1, $2, $3) RETURNING id
		`,
		user.Name,
		user.Email,
		user.Password,
	)

	err := row.Scan(&user.ID)

	if err != nil {
		return models.User{}, fmt.Errorf(
			"[in services.UsersService.CreateUser] failed to create user: %w",
			err,
		)
	}

	return user, nil
}

// ReadUser attempts to read a user from the database using the provided id. A
// fully hydrated models.User or error is returned.
func (s *UsersService) ReadUser(ctx context.Context, id uint64) (models.User, error) {
	s.logger.DebugContext(ctx, "Reading user", "id", id)

	row := s.db.QueryRowContext(
		ctx,
		`
		SELECT id,
		       name,
		       email,
		       password
		FROM users
		WHERE id = $1::int
        `,
		id,
	)

	var user models.User

	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return models.User{}, nil
		default:
			return models.User{}, fmt.Errorf(
				"[in services.UsersService.ReadUser] failed to read user: %w",
				err,
			)
		}
	}

	return user, nil
}

// UpdateUser attempts to perform an update of the user with the provided id,
// updating, it to reflect the properties on the provided patch object. A
// models.User or an error.
func (s *UsersService) UpdateUser(ctx context.Context, id uint64, patch models.User) (models.User, error) {
	s.logger.DebugContext(ctx, "Updating user", "id", id)

	_, err := s.db.ExecContext(
		ctx,
		`
		UPDATE users 
		SET name = $1, email = $2, password = $3
		WHERE id = $4
		`,
		patch.Name,
		patch.Email,
		patch.Password,
		id,
	)

	if err != nil {
		return models.User{}, fmt.Errorf(
			"[in services.UsersService.UpdateUser] failed to update user: %w",
			err,
		)
	}
	patch.ID = uint(id)
	return patch, nil
}

// DeleteUser attempts to delete the user with the provided id. An error is
// returned if the delete fails.
func (s *UsersService) DeleteUser(ctx context.Context, id uint64) error {
	s.logger.DebugContext(ctx, "Deleting user", "id", id)

	// Delete user from user table
	_, err := s.db.ExecContext(
		ctx,
		`
		DELETE FROM users WHERE id = $1::int
		`,
		id,
	)

	if err != nil {
		return fmt.Errorf(
			"[in services.UsersService.DeleteUser] failed to delete user: %w",
			err,
		)
	}

	// Delete blogs with authodId = id
	_, err = s.db.ExecContext(
		ctx,
		`
		DELETE FROM blogs WHERE author_id = $1::int
		`,
		id,
	)

	if err != nil {
		return fmt.Errorf(
			"[in services.UsersService.DeleteUser] failed to delete blogs: %w",
			err,
		)
	}

	//Delete comments where userId = id

	_, err = s.db.ExecContext(
		ctx,
		`
		DELETE FROM comments WHERE user_id = $1::int
		`,
		id,
	)

	if err != nil {
		return fmt.Errorf(
			"[in services.UsersService.DeleteUser] failed to delete comments: %w",
			err,
		)
	}

	return nil
}

// ListUsers attempts to list all users in the database. A slice of models.User
// or an error is returned.
func (s *UsersService) ListUsers(ctx context.Context, name string) ([]models.User, error) {
	s.logger.DebugContext(ctx, "Listing users")

	rows, err := s.db.QueryContext(
		ctx,
		`
		SELECT id,
		       name,
		       email,
		       password
		FROM users
        `,
	)

	if err != nil {
		return []models.User{}, fmt.Errorf(
			"[in services.UsersService.ListUser] failed to read users: %w",
			err,
		)
	}

	var users []models.User

	// error handling for error in for loop and then for empty error out of for loop?
	// NEED TO HANDLE ERROR FOR EMPTY RESULTS
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Password)
		if err != nil {
			return []models.User{}, fmt.Errorf(
				"[in services.UsersService.ListUser] failed to read users: %w",
				err,
			)
		}

		if name == "" || user.Name == name {
			users = append(users, user)
		}
	}

	if err = rows.Err(); err != nil {
		return []models.User{}, fmt.Errorf(
			"[in services.UsersService.ListUser] failed to read users: %w",
			err,
		)
	}

	return users, nil
}
