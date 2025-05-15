package models

import (
	"database/sql"
	"errors"
)

type User struct {
	ID int
	Username string
	Email string
	Password string
}

// Insert new user
func CreateUser(username, email, hashedPassword string) error {
	_, err := DB.Exec("INSERT INTO users (username, email, password) VALUES ($1, $2, $3)",
		username, email, hashedPassword)
	return err
}

// Fetch user by email
func GetUserByEmail(email string) (User, error) {
	var user User
	row := DB.QueryRow("SELECT id, username, email, password FROM users WHERE email = $1", email)
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return user, errors.New("user not found")
		}
		return user, err
	}
	return user, nil
}