package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
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

	server, err := NewTestServer()
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

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("Expected status 200 OK, got %d", resp.StatusCode)
			}

			var user struct {
				ID       int    `json:"id"`
				Name     string `json:"name"`
				Email    string `json:"email"`
				Password string `json:"password"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if user.ID != tc.id || user.Name != tc.name || user.Email != tc.email || user.Password != tc.password {
				t.Errorf("Unexpected user data: %+v", user)
			}
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

	server, err := NewTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	t.Cleanup(server.Close)

	resp, err := http.Get(server.URL + "/api/users")
	if err != nil {
		t.Fatalf("Failed to make GET request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200 OK, got %d", resp.StatusCode)
	}

	var users []struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(users) != len(expected) {
		t.Fatalf("Expected %d users, got %d", len(expected), len(users))
	}

	for i, exp := range expected {
		got := users[i]
		if got.ID != exp.ID || got.Name != exp.Name || got.Email != exp.Email || got.Password != exp.Password {
			t.Errorf("User %d: expected %+v, got %+v", i, exp, got)
		}
	}
}

func TestCreateUser(t *testing.T) {
	t.Parallel()

	server, err := NewTestServer()
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

	var createdUser struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&createdUser); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if createdUser.Name != newUser.Name || createdUser.Email != newUser.Email || createdUser.Password != newUser.Password {
		t.Errorf("Unexpected user data: %+v", createdUser)
	}

	// check that the user is now in the list
	resp2, err := http.Get(server.URL + "/api/users")
	if err != nil {
		t.Fatalf("Failed to make GET request: %v", err)
	}
	defer resp2.Body.Close()

	var users []struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(resp2.Body).Decode(&users); err != nil {
		t.Fatalf("Failed to decode users list: %v", err)
	}

	found := false
	for _, u := range users {
		if u.Email == newUser.Email && u.Name == newUser.Name && u.Password == newUser.Password {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Created user not found in users list")
	}
}

func TestUpdateUser(t *testing.T) {
	t.Parallel()

	server, err := NewTestServer()
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

	req, err := http.NewRequest(http.MethodPut, server.URL+"/api/user/1", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to create PUT request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to make PUT request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200 OK, got %d", resp.StatusCode)
	}

	var user struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if user.ID != 1 || user.Name != updatedUser.Name || user.Email != updatedUser.Email || user.Password != updatedUser.Password {
		t.Errorf("Unexpected user data after update: %+v", user)
	}

	// perform a GET request for ID 1 and check the updated information
	getResp, err := http.Get(server.URL + "/api/user/1")
	if err != nil {
		t.Fatalf("Failed to make GET request: %v", err)
	}
	defer getResp.Body.Close()

	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200 OK, got %d", getResp.StatusCode)
	}

	var gotUser struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(getResp.Body).Decode(&gotUser); err != nil {
		t.Fatalf("Failed to decode GET response: %v", err)
	}

	if gotUser.ID != 1 || gotUser.Name != updatedUser.Name || gotUser.Email != updatedUser.Email || gotUser.Password != updatedUser.Password {
		t.Errorf("GET after update: unexpected user data: %+v", gotUser)
	}
}

func TestDeleteUser(t *testing.T) {
	t.Parallel()

	server, err := NewTestServer()
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	t.Cleanup(server.Close)

	// Delete user with ID 2 (Bob)
	req, err := http.NewRequest(http.MethodDelete, server.URL+"/api/user/2", nil)
	if err != nil {
		t.Fatalf("Failed to create DELETE request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to make DELETE request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("Expected status 204 No Content, got %d", resp.StatusCode)
	}

	// Get the list of users after deletion
	listResp, err := http.Get(server.URL + "/api/users")
	if err != nil {
		t.Fatalf("Failed to make GET request: %v", err)
	}
	defer listResp.Body.Close()

	if listResp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200 OK, got %d", listResp.StatusCode)
	}

	var users []struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(listResp.Body).Decode(&users); err != nil {
		t.Fatalf("Failed to decode users list: %v", err)
	}

	// Ensure user with ID 2 (Bob) is not in the list
	for _, u := range users {
		if u.ID == 2 || u.Name == "Bob" || u.Email == "bob@example.com" {
			t.Errorf("Deleted user still found in users list: %+v", u)
		}
	}
}
