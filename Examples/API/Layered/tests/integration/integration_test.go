package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadUser(t *testing.T) {
	t.Parallel()
	tests := []struct {
		id       int
		name     string
		email    string
		password string
	}{
		{1, "Alice", "alice@example.com", "password123"},
		{2, "Bob", "bob@example.com", "securepass456"},
		{3, "Carol", "carol@example.com", "carolpass789"},
		{4, "Dave", "dave@example.com", "davepass321"},
	}

	server, _, err := newTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	t.Cleanup(server.Close)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := http.Get(server.URL + "/api/user/" + strconv.Itoa(tc.id))
			if err != nil {
				t.Fatalf("Failed to make GET request: %v", err)
			}
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status code 200 OK")

			var user struct {
				ID    int    `json:"id"`
				Name  string `json:"name"`
				Email string `json:"email"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			assert.Equal(t, tc.id, user.ID, "User ID mismatch")
			assert.Equal(t, tc.name, user.Name, "User name mismatch")
			assert.Equal(t, tc.email, user.Email, "User email mismatch")
		})
	}
}

func TestListUsers(t *testing.T) {
	t.Parallel()

	expected := []struct {
		ID       int
		Name     string
		Email    string
		Password string
	}{
		{1, "Alice", "alice@example.com", "password123"},
		{2, "Bob", "bob@example.com", "securepass456"},
		{3, "Carol", "carol@example.com", "carolpass789"},
		{4, "Dave", "dave@example.com", "davepass321"},
	}

	server, _, err := newTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	t.Cleanup(server.Close)

	resp, err := http.Get(server.URL + "/api/user")
	if err != nil {
		t.Fatalf("Failed to make GET request: %v", err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status code 200 OK")

	var response struct {
		Users []struct {
			ID    int    `json:"id"`
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"Users"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	assert.Len(t, response.Users, len(expected), "Expected number of users does not match")

	for i, exp := range expected {
		got := response.Users[i]
		assert.Equal(t, exp.ID, got.ID, "User ID mismatch at index %d", i)
		assert.Equal(t, exp.Name, got.Name, "User name mismatch at index %d", i)
		assert.Equal(t, exp.Email, got.Email, "User email mismatch at index %d", i)
	}
}

func TestCreateUser(t *testing.T) {
	t.Parallel()

	server, db, err := newTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	t.Cleanup(server.Close)

	newUser := struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}{
		Name:     "Eve",
		Email:    "eve@example.com",
		Password: "evepass999",
	}

	// Marshal new user to JSON
	body, err := json.Marshal(newUser)
	if err != nil {
		t.Fatalf("Failed to marshal user: %v", err)
	}

	resp, err := http.Post(server.URL+"/api/user", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to make POST request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status 201 Created, got %d", resp.StatusCode)
	}

	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Expected status code 201 Created")

	var createdUser struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&createdUser); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	assert.Equal(t, newUser.Name, createdUser.Name, "Created user name mismatch")
	assert.Equal(t, newUser.Email, createdUser.Email, "Created user email mismatch")

	// Verify the new user was inserted into the database
	var dbUser struct {
		ID       int    `db:"id"`
		Name     string `db:"name"`
		Email    string `db:"email"`
		Password string `db:"password"`
	}
	err = db.Get(
		&dbUser,
		"SELECT id, name, email, password FROM users WHERE email = ?",
		newUser.Email,
	)
	if err != nil {
		t.Fatalf("Failed to query user from DB: %v", err)
	}
	assert.Equal(t, newUser.Name, dbUser.Name, "DB user name mismatch")
	assert.Equal(t, newUser.Email, dbUser.Email, "DB user email mismatch")
	assert.Equal(t, newUser.Password, dbUser.Password, "DB user password mismatch")
	assert.Equal(t, createdUser.ID, dbUser.ID, "DB user ID mismatch with API response")
}

func TestUpdateUser(t *testing.T) {
	t.Parallel()

	server, db, err := newTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	t.Cleanup(server.Close)

	updatedUser := struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}{
		Name:     "Alice Updated",
		Email:    "alice.updated@example.com",
		Password: "newpassword123",
	}

	// Marshal updated user to JSON
	body, err := json.Marshal(updatedUser)
	if err != nil {
		t.Fatalf("Failed to marshal user: %v", err)
	}

	req, err := http.NewRequestWithContext(
		t.Context(),
		http.MethodPut,
		server.URL+"/api/user/1",
		bytes.NewReader(body),
	)
	if err != nil {
		t.Fatalf("Failed to create PUT request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to make PUT request: %v", err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status code 200 OK")

	var user struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	assert.Equal(t, updatedUser.Name, user.Name, "Updated user name mismatch")
	assert.Equal(t, updatedUser.Email, user.Email, "Updated user email mismatch")
	assert.Equal(t, 1, user.ID, "Updated user ID mismatch")

	// Verify the user was updated in the database
	var dbUser struct {
		ID       int    `db:"id"`
		Name     string `db:"name"`
		Email    string `db:"email"`
		Password string `db:"password"`
	}
	err = db.Get(&dbUser, "SELECT id, name, email, password FROM users WHERE id = ?", 1)
	if err != nil {
		t.Fatalf("Failed to query user from DB: %v", err)
	}
	assert.Equal(t, updatedUser.Name, dbUser.Name, "DB user name mismatch after update")
	assert.Equal(t, updatedUser.Email, dbUser.Email, "DB user email mismatch after update")
	assert.Equal(t, updatedUser.Password, dbUser.Password, "DB user password mismatch after update")
	assert.Equal(t, 1, dbUser.ID, "DB user ID mismatch after update")
}

func TestDeleteUser(t *testing.T) {
	t.Parallel()

	server, db, err := newTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	t.Cleanup(server.Close)

	// Delete user with ID 2 (Bob)
	req, err := http.NewRequestWithContext(
		t.Context(),
		http.MethodDelete,
		server.URL+"/api/user/2",
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create DELETE request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to make DELETE request: %v", err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode, "Expected status code 204 No Content")

	// Verify user is no longer in the database
	var dbUser struct {
		ID       int    `db:"id"`
		Name     string `db:"name"`
		Email    string `db:"email"`
		Password string `db:"password"`
	}
	err = db.Get(&dbUser, "SELECT id, name, email, password FROM users WHERE id = ?", 2)
	assert.Error(t, err, "Expected error when querying deleted user")
}
