package services

import (
	"context"
	"database/sql/driver"
	"log/slog"
	"regexp"
	"testing"

	"example.com/examples/api/layered/models"
	"github.com/DATA-DOG/go-sqlmock"
)

func TestUsersService_ReadUser(t *testing.T) {
	testcases := map[string]struct {
		mockCalled     bool
		mockInputArgs  []driver.Value
		mockOutput     *sqlmock.Rows
		mockError      error
		input          uint64
		expectedOutput models.User
		expectedError  error
	}{
		"happy path": {
			mockCalled:    true,
			mockInputArgs: []driver.Value{1},
			mockOutput: sqlmock.NewRows([]string{"id", "name", "email", "password"}).
				AddRow(1, "john", "john@me.com", "password123!"),
			mockError: nil,
			input:     1,
			expectedOutput: models.User{
				ID:       1,
				Name:     "john",
				Email:    "john@me.com",
				Password: "password123!",
			},
			expectedError: nil,
		},
	}
	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			logger := slog.Default()

			if tc.mockCalled {
				mock.
					ExpectQuery(regexp.QuoteMeta(`
                        SELECT id,
                               name,
                               email,
                               password
                        FROM users
                        WHERE id = $1::int
                    `)).
					WithArgs(tc.mockInputArgs...).
					WillReturnRows(tc.mockOutput).
					WillReturnError(tc.mockError)
			}

			userService := NewUsersService(logger, db)

			output, err := userService.ReadUser(context.TODO(), tc.input)
			if err != tc.expectedError {
				t.Errorf("expected no error, got %v", err)
			}
			if output != tc.expectedOutput {
				t.Errorf("expected %v, got %v", tc.expectedOutput, output)
			}

			if tc.mockCalled {
				if err = mock.ExpectationsWereMet(); err != nil {
					t.Errorf("there were unfulfilled expectations: %s", err)
				}
			}
		})
	}
}
func TestUsersService_ListUsers(t *testing.T) {
	testcases := map[string]struct {
		mockCalled     bool
		mockInputArgs  []driver.Value
		mockOutput     *sqlmock.Rows
		mockError      error
		input          string
		expectedOutput []models.User
		expectedError  error
	}{
		"happy path": {
			mockCalled:    true,
			mockInputArgs: []driver.Value{1},
			mockOutput: sqlmock.NewRows([]string{"id", "name", "email", "password"}).
				AddRow(1, "john", "john@me.com", "password123!").
				AddRow(2, "jane", "jane@me.com", "pwd5678!"),

			mockError: nil,
			input:     "",
			expectedOutput: []models.User{
				{
					ID:       1,
					Name:     "john",
					Email:    "john@me.com",
					Password: "password123!",
				},
				{
					ID:       2,
					Name:     "jane",
					Email:    "jane@me.com",
					Password: "pwd5678!",
				},
			},
			expectedError: nil,
		},
		"filter by name": {
			mockCalled:    true,
			mockInputArgs: []driver.Value{1},
			mockOutput: sqlmock.NewRows([]string{"id", "name", "email", "password"}).
				AddRow(1, "john", "john@me.com", "password123!").
				AddRow(2, "jane", "jane@me.com", "pwd5678!"),

			mockError: nil,
			input:     "jane",
			expectedOutput: []models.User{
				{
					ID:       2,
					Name:     "jane",
					Email:    "jane@me.com",
					Password: "pwd5678!",
				},
			},
			expectedError: nil,
		},
		"filter by name no results": {
			mockCalled:    true,
			mockInputArgs: []driver.Value{1},
			mockOutput: sqlmock.NewRows([]string{"id", "name", "email", "password"}).
				AddRow(1, "john", "john@me.com", "password123!").
				AddRow(2, "jane", "jane@me.com", "pwd5678!"),

			mockError:      nil,
			input:          "joe",
			expectedOutput: []models.User{},
			expectedError:  nil,
		},
	}
	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			logger := slog.Default()

			if tc.mockCalled {
				mock.
					ExpectQuery(regexp.QuoteMeta(`
                        SELECT id,
							name,
							email,
							password
						FROM users
                    `)).
					WillReturnRows(tc.mockOutput).
					WillReturnError(tc.mockError)
			}

			userService := NewUsersService(logger, db)

			outputs, err := userService.ListUsers(context.TODO(), tc.input)
			if err != tc.expectedError {
				t.Errorf("expected no error, got %v", err)
			}

			for i, output := range outputs {
				if output != tc.expectedOutput[i] {
					t.Errorf("expected %v, got %v", tc.expectedOutput[i], output)
				}
			}

			if tc.mockCalled {
				if err = mock.ExpectationsWereMet(); err != nil {
					t.Errorf("there were unfulfilled expectations: %s", err)
				}
			}
		})
	}
}

func TestBlogsService_DeleteUser(t *testing.T) {
	testcases := map[string]struct {
		mockCalled     bool
		mockInputArgs  []driver.Value
		mockOutput     *sqlmock.Rows
		mockError      error
		input          uint64
		expectedOutput models.Blog
		expectedError  error
	}{
		"happy path": {
			mockCalled:    true,
			mockInputArgs: []driver.Value{1},
			mockError:     nil,
			input:         1,
			expectedError: nil,
		},
	}
	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			logger := slog.Default()

			if tc.mockCalled {
				mock.
					ExpectExec(regexp.QuoteMeta(`DELETE FROM users WHERE id = $1::int`)).
					WithArgs(tc.mockInputArgs...).
					WillReturnResult(sqlmock.NewResult(1, 1)).
					WillReturnError(tc.mockError)
				mock.
					ExpectExec(regexp.QuoteMeta(`DELETE FROM blogs WHERE author_id = $1::int`)).
					WithArgs(tc.mockInputArgs...).
					WillReturnResult(sqlmock.NewResult(1, 1)).
					WillReturnError(tc.mockError)

				mock.
					ExpectExec(regexp.QuoteMeta(`DELETE FROM comments WHERE user_id = $1::int`)).
					WithArgs(tc.mockInputArgs...).
					WillReturnResult(sqlmock.NewResult(1, 1)).
					WillReturnError(tc.mockError)
			}

			userService := NewUsersService(logger, db)

			err = userService.DeleteUser(context.TODO(), tc.input)
			if err != tc.expectedError {
				t.Errorf("expected no error, got %v", err)
			}

			if tc.mockCalled {
				if err = mock.ExpectationsWereMet(); err != nil {
					t.Errorf("there were unfulfilled expectations: %s", err)
				}
			}
		})
	}
}

func TestUsersService_CreateUser(t *testing.T) {
	testcases := map[string]struct {
		mockCalled     bool
		mockInputArgs  []driver.Value
		mockOutput     *sqlmock.Rows
		mockError      error
		input          models.User
		expectedOutput models.User
		expectedError  error
	}{
		"happy path": {
			mockCalled:    true,
			mockInputArgs: []driver.Value{"john", "john@me.com", "password123!"},
			mockOutput: sqlmock.NewRows([]string{"id"}).
				AddRow(1),
			mockError: nil,
			input: models.User{
				Name:     "john",
				Email:    "john@me.com",
				Password: "password123!",
			},
			expectedOutput: models.User{
				ID:       1,
				Name:     "john",
				Email:    "john@me.com",
				Password: "password123!",
			},
			expectedError: nil,
		},
	}
	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			logger := slog.Default()

			if tc.mockCalled {
				mock.
					ExpectQuery(regexp.QuoteMeta(`
                        INSERT INTO users (name, email, password) VALUES ($1, $2, $3) RETURNING id
                    `)).
					WithArgs(tc.mockInputArgs...).
					WillReturnRows(tc.mockOutput).
					WillReturnError(tc.mockError)
			}

			userService := NewUsersService(logger, db)

			output, err := userService.CreateUser(context.TODO(), tc.input)
			if err != tc.expectedError {
				t.Errorf("expected no error, got %v", err)
			}
			if output != tc.expectedOutput {
				t.Errorf("expected %v, got %v", tc.expectedOutput, output)
			}

			if tc.mockCalled {
				if err = mock.ExpectationsWereMet(); err != nil {
					t.Errorf("there were unfulfilled expectations: %s", err)
				}
			}
		})
	}
}

func TestUsersService_UpdateUser(t *testing.T) {
	testcases := map[string]struct {
		mockCalled     bool
		mockInputArgs  []driver.Value
		mockOutput     *sqlmock.Rows
		mockError      error
		input          models.User
		expectedOutput models.User
		expectedError  error
	}{
		"happy path": {
			mockCalled:    true,
			mockInputArgs: []driver.Value{"john", "john@me.com", "password123!", 1},
			mockOutput: sqlmock.NewRows([]string{"id"}).
				AddRow(1),
			mockError: nil,
			input: models.User{
				Name:     "john",
				Email:    "john@me.com",
				Password: "password123!",
			},
			expectedOutput: models.User{
				ID:       1,
				Name:     "john",
				Email:    "john@me.com",
				Password: "password123!",
			},
			expectedError: nil,
		},
	}
	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			logger := slog.Default()

			if tc.mockCalled {
				mock.
					ExpectExec(regexp.QuoteMeta(`
                        UPDATE users 
						SET name = $1, email = $2, password = $3
						WHERE id = $4
                    `)).
					WithArgs(tc.mockInputArgs...).
					WillReturnResult(sqlmock.NewResult(1, 1)).
					WillReturnError(tc.mockError)
			}

			userService := NewUsersService(logger, db)

			output, err := userService.UpdateUser(context.TODO(), 1, tc.input)
			if err != tc.expectedError {
				t.Errorf("expected no error, got %v", err)
			}
			if output != tc.expectedOutput {
				t.Errorf("expected %v, got %v", tc.expectedOutput, output)
			}

			if tc.mockCalled {
				if err = mock.ExpectationsWereMet(); err != nil {
					t.Errorf("there were unfulfilled expectations: %s", err)
				}
			}
		})
	}
}
