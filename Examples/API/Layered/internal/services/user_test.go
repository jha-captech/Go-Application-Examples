package services

import (
	"context"
	"database/sql/driver"
	"errors"
	"log/slog"
	"regexp"
	"strconv"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-redis/redismock/v9"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"example.com/examples/api/layered/internal/models"
)

func TestUsersService_DeepHealthCheck(t *testing.T) {
	type fields struct {
		dbErr    error
		cacheErr error
	}
	tests := map[string]struct {
		fields      fields
		wantStatus  []HealthStatus
		wantErr     bool
		errContains string
	}{
		"healthy": {
			fields: fields{dbErr: nil, cacheErr: nil},
			wantStatus: []HealthStatus{
				{Name: "db", Status: "healthy"},
				{Name: "cache", Status: "healthy"},
			},
			wantErr:     false,
			errContains: "",
		},
		"db ping error": {
			fields: fields{dbErr: errors.New("db down"), cacheErr: nil},
			wantStatus: []HealthStatus{
				{Name: "db", Status: "unhealthy"},
				{Name: "cache", Status: "healthy"},
			},
			wantErr:     true,
			errContains: "failed to ping database",
		},
		"cache ping error": {
			fields: fields{dbErr: nil, cacheErr: errors.New("cache down")},
			wantStatus: []HealthStatus{
				{Name: "db", Status: "healthy"},
				{Name: "cache", Status: "unhealthy"},
			},
			wantErr:     true,
			errContains: "failed to ping cache",
		},
		"both ping error": {
			fields: fields{dbErr: errors.New("db down"), cacheErr: errors.New("cache down")},
			wantStatus: []HealthStatus{
				{Name: "db", Status: "unhealthy"},
				{Name: "cache", Status: "unhealthy"},
			},
			wantErr:     true,
			errContains: "failed to ping database",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
			assert.NoError(t, err)
			defer db.Close()

			// Mock DB ping
			mock.ExpectPing().WillReturnError(tc.fields.dbErr)

			// Mock cache ping
			rdb, rmock := redismock.NewClientMock()
			if tc.fields.cacheErr == nil {
				rmock.ExpectPing().SetVal("OK")
			} else {
				rmock.ExpectPing().SetErr(tc.fields.cacheErr)
			}

			logger := slog.Default()
			us := NewUsersService(logger, sqlx.NewDb(db, "sqlmock"), rdb, 0)

			status, err := us.DeepHealthCheck(context.Background())
			assert.Equal(t, tc.wantStatus, status)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
			assert.NoError(t, rmock.ExpectationsWereMet())
		})
	}
}

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
				t.Fatalf(
					"an error '%s' was not expected when opening a stub database connection",
					err,
				)
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

			rdb, rmock := redismock.NewClientMock()
			rmock.ExpectGet(strconv.FormatUint(tc.input, 10)).SetErr(redis.Nil)
			rmock.Regexp().ExpectSet(strconv.Itoa(int(tc.expectedOutput.ID)), `.*`, 0).SetVal("OK")
			userService := NewUsersService(logger, sqlx.NewDb(db, "sqlmock"), rdb, 0)

			output, err := userService.ReadUser(t.Context(), tc.input)
			require.ErrorIs(t, err, tc.expectedError)
			assert.Equal(t, tc.expectedOutput, output)

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
				AddRow(2, "jane", "jane@me.com", "pwd5678!"),

			mockError: nil,
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
			mockOutput:    sqlmock.NewRows([]string{"id", "name", "email", "password"}),

			mockError:      nil,
			expectedOutput: []models.User{},
			expectedError:  nil,
		},
	}
	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf(
					"an error '%s' was not expected when opening a stub database connection",
					err,
				)
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

			rdb, _ := redismock.NewClientMock()
			userService := NewUsersService(logger, sqlx.NewDb(db, "sqlmock"), rdb, 0)

			outputs, err := userService.ListUsers(t.Context())
			require.ErrorIs(t, err, tc.expectedError)

			for i, output := range outputs {
				assert.Equal(t, tc.expectedOutput[i], output)
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
		mockCalled    bool
		mockInputArgs []driver.Value
		mockOutput    *sqlmock.Rows
		mockError     error
		input         uint64
		expectedError error
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
				t.Fatalf(
					"an error '%s' was not expected when opening a stub database connection",
					err,
				)
			}
			defer db.Close()

			logger := slog.Default()

			if tc.mockCalled {
				mock.
					ExpectExec(regexp.QuoteMeta(`DELETE FROM users WHERE id = $1::int`)).
					WithArgs(tc.mockInputArgs...).
					WillReturnResult(sqlmock.NewResult(1, 1)).
					WillReturnError(tc.mockError)
			}

			rdb, rmock := redismock.NewClientMock()
			rmock.ExpectDel(strconv.FormatUint(tc.input, 10)).SetVal(1)
			userService := NewUsersService(logger, sqlx.NewDb(db, "sqlmock"), rdb, 0)

			err = userService.DeleteUser(t.Context(), tc.input)
			assert.ErrorIs(t, err, tc.expectedError)

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
				t.Fatalf(
					"an error '%s' was not expected when opening a stub database connection",
					err,
				)
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

			rdb, rmock := redismock.NewClientMock()
			rmock.Regexp().ExpectSet(strconv.Itoa(int(tc.expectedOutput.ID)), `.*`, 0).SetVal("OK")
			userService := NewUsersService(logger, sqlx.NewDb(db, "sqlmock"), rdb, 0)

			output, err := userService.CreateUser(t.Context(), tc.input)
			assert.ErrorIs(t, err, tc.expectedError)
			assert.Equal(t, tc.expectedOutput, output)

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
			mockOutput: sqlmock.NewRows([]string{"id", "name", "email", "password"}).
				AddRow(1, "john", "john@me.com", "password123!"),
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
				t.Fatalf(
					"an error '%s' was not expected when opening a stub database connection",
					err,
				)
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

				mock.
					ExpectQuery(regexp.QuoteMeta(`
                        SELECT id,
                               name,
                               email,
                               password
                        FROM users
                        WHERE id = $1::int
                    `)).
					WithArgs(tc.expectedOutput.ID).
					WillReturnRows(tc.mockOutput).
					WillReturnError(tc.mockError)
			}

			rdb, rmock := redismock.NewClientMock()
			rmock.ExpectGet(strconv.FormatUint(uint64(tc.expectedOutput.ID), 10)).SetErr(redis.Nil)
			rmock.Regexp().ExpectSet(strconv.Itoa(int(tc.expectedOutput.ID)), `.*`, 0).SetVal("OK")
			rmock.Regexp().ExpectSet(strconv.Itoa(int(tc.expectedOutput.ID)), `.*`, 0).SetVal("OK")
			userService := NewUsersService(logger, sqlx.NewDb(db, "sqlmock"), rdb, 0)

			output, err := userService.UpdateUser(t.Context(), 1, tc.input)
			assert.ErrorIs(t, err, tc.expectedError)
			assert.Equal(t, tc.expectedOutput, output)

			if tc.mockCalled {
				if err = mock.ExpectationsWereMet(); err != nil {
					t.Errorf("there were unfulfilled expectations: %s", err)
				}
			}
		})
	}
}
