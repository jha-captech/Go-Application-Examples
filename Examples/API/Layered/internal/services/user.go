package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"example.com/examples/api/layered/internal/models"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

// UsersService is a service capable of performing CRUD operations for
// models.User models.
type UsersService struct {
	logger *slog.Logger
	db     *sqlx.DB
	cache  *Client
}

// NewUsersService creates a new UsersService and returns a pointer to it.
func NewUsersService(logger *slog.Logger, db *sqlx.DB, rdb *redis.Client, expiration time.Duration) *UsersService {
	return &UsersService{
		logger: logger,
		db:     db,
		cache:  NewClient(rdb, expiration),
	}
}

// CreateUser attempts to create the provided user, returning a fully hydrated
// models.User or an error.
func (s *UsersService) CreateUser(ctx context.Context, user models.User) (models.User, error) {
	logger := s.logger.With(slog.String("func", "services.UsersService.CreateUser"))
	logger.DebugContext(ctx, "Creating user", "name", user.Name)

	err := s.db.QueryRowContext(
		ctx,
		`
		INSERT 
		INTO users (name, email, password) 
		VALUES ($1, $2, $3) 
		RETURNING id
		`,
		user.Name,
		user.Email,
		user.Password,
	).Scan(&user.ID)
	if err != nil {
		return models.User{}, fmt.Errorf(
			"[in services.UsersService.CreateUser] failed to create user: %w",
			err,
		)
	}

	// Write the user to the cache
	logger.DebugContext(ctx, "Setting user in cache", "id", user.ID)
	if err = s.cache.SetMarshal(ctx, strconv.Itoa(int(user.ID)), user); err != nil {
		return models.User{}, fmt.Errorf("[in services.UsersService.CreateUser] failed to write user to cache: %w", err)
	}

	return user, nil
}

// ReadUser attempts to read a user from the database using the provided id. A
// fully hydrated models.User or error is returned.
func (s *UsersService) ReadUser(ctx context.Context, id uint64) (models.User, error) {
	logger := s.logger.With(slog.String("func", "services.UsersService.ReadUser"))
	logger.DebugContext(ctx, "Reading user", "id", id)

	// Check the cache for the user object
	logger.DebugContext(ctx, "Reading user from cache", "id", id)

	var user models.User
	found, err := s.cache.Get(ctx, strconv.FormatUint(id, 10)).Unmarshal(&user)
	if err != nil {
		return models.User{}, fmt.Errorf(
			"[in services.UsersService.ReadUser] failed to read user from cache: %w",
			err,
		)
	}

	// If the user was found in the cache, return it
	if found {
		return user, nil
	}

	err = s.db.GetContext(
		ctx,
		&user,
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

	// Write the user to the cache
	logger.DebugContext(ctx, "Setting user in cache", "id", id)
	if err = s.cache.SetMarshal(ctx, strconv.FormatUint(id, 10), user); err != nil {
		return models.User{}, fmt.Errorf(
			"[in services.UsersService.ReadUser] failed to write user to cache: %w",
			err,
		)
	}

	return user, nil
}

// UpdateUser attempts to perform an update of the user with the provided id,
// updating, it to reflect the properties on the provided patch object. A
// models.User or an error.
func (s *UsersService) UpdateUser(ctx context.Context, id uint64, patch models.User) (models.User, error) {
	logger := s.logger.With(slog.String("func", "services.UsersService.UpdateUser"))
	logger.DebugContext(ctx, "Updating user", "id", id, "patch", patch)

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

	// Read the updated user from the database
	user, err := s.ReadUser(ctx, id)
	if err != nil {
		return models.User{}, fmt.Errorf("[in services.UsersService.UpdateUser] failed to read updated user: %w", err)
	}

	// Write the updated user to the cache
	logger.DebugContext(ctx, "Setting updated user in cache", "id", id)
	if err = s.cache.SetMarshal(ctx, strconv.FormatUint(id, 10), user); err != nil {
		return models.User{}, fmt.Errorf("[in services.UsersService.UpdateUser] failed to write updated user to cache: %w", err)
	}

	patch.ID = uint(id)
	return patch, nil
}

// DeleteUser attempts to delete the user with the provided id. An error is
// returned if the delete fails.
func (s *UsersService) DeleteUser(ctx context.Context, id uint64) error {
	logger := s.logger.With(slog.String("func", "services.UsersService.DeleteUser"))
	logger.DebugContext(ctx, "Deleting user", "id", id)

	// Delete user from user table
	_, err := s.db.ExecContext(
		ctx,
		`
		DELETE 
		FROM users 
		WHERE id = $1::int
		`,
		id,
	)
	if err != nil {
		return fmt.Errorf(
			"[in services.UsersService.DeleteUser] failed to delete user: %w",
			err,
		)
	}

	// Remove the user from the cache
	logger.DebugContext(ctx, "Removing user from cache", "id", id)
	if err = s.cache.Del(ctx, strconv.FormatUint(id, 10)).Err(); err != nil {
		return fmt.Errorf("[in services.UsersService.DeleteUser] failed to remove user from cache: %w", err)
	}

	return nil
}

// ListUsers attempts to list all users in the database. A slice of models.User
// or an error is returned.
func (s *UsersService) ListUsers(ctx context.Context, name string) ([]models.User, error) {
	s.logger.DebugContext(ctx, "Listing users")
	var users []models.User

	err := s.db.SelectContext(
		ctx,
		&users,
		`
		SELECT id,
		       name,
		       email,
		       password
		FROM users
		WHERE name = $1::text
        `,
		name,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"[in services.UsersService.ListUser] failed to read users: %w",
			err,
		)
	}

	return users, nil
}
