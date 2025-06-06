package models

// User represents a user in the system.
type User struct {
	ID       uint   `db:"id"       json:"id"`
	Name     string `db:"name"     json:"name"`
	Email    string `db:"email"    json:"email"`
	Password string `db:"password" json:"password"`
}
