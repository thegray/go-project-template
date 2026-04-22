package order

import "context"

type Repository interface {
	Create(ctx context.Context, order *Order) error
	GetByID(ctx context.Context, id string) (*Order, error)
	ListByUser(ctx context.Context, userID string, page, limit int) ([]Order, error)
	Update(ctx context.Context, order *Order) error
	Delete(ctx context.Context, id string) error
}
