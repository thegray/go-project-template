package rest

import (
	"time"

	"project-template/internal/user"
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type UserPathParam struct {
	ID string `uri:"id" binding:"required,uuid"`
}

type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func NewUserResponse(u *user.User) UserResponse {
	if u == nil {
		return UserResponse{}
	}

	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
	}
}
