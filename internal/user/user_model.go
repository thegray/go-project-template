package user

import "time"

type User struct {
	ID           string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
