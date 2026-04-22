package checkout

import (
	"context"

	"project-template/internal/order"
	"project-template/internal/user"
)

type UserProvider interface {
	GetByID(ctx context.Context, id string) (*user.User, error)
}

type OrderCreator interface {
	Create(ctx context.Context, input order.CreateInput) (*order.Order, error)
}

type Input struct {
	UserID string       `json:"user_id"`
	Items  []order.Item `json:"items"`
}

type Result struct {
	User          *user.User
	Order         *order.Order
	DiscountGiven int
}
