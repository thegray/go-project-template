package user

import "context"

type Repository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) error
}

type ServiceInterface interface {
	Login(ctx context.Context, input LoginInput) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
}
