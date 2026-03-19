package order

import "time"

type Item struct {
	SKU       string `json:"sku"`
	Name      string `json:"name"`
	Quantity  int    `json:"quantity"`
	UnitPrice int64  `json:"unit_price"`
}

type Order struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Status    string    `json:"status"`
	Items     []Item    `json:"items"`
	Subtotal  int64     `json:"subtotal"`
	Discount  int64     `json:"discount"`
	Total     int64     `json:"total"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateInput struct {
	UserID          string `json:"user_id"`
	Status          string `json:"status"`
	DiscountPercent int    `json:"discount_percent"`
	Items           []Item `json:"items"`
}

type UpdateInput struct {
	UserID          string `json:"user_id"`
	Status          string `json:"status"`
	DiscountPercent int    `json:"discount_percent"`
	Items           []Item `json:"items"`
}
