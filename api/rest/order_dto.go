package rest

import (
	"time"

	"project-template/internal/order"
)

type OrderItemRequest struct {
	SKU       string `json:"sku" binding:"required"`
	Name      string `json:"name" binding:"required"`
	Quantity  int    `json:"quantity" binding:"min=1"`
	UnitPrice int64  `json:"unit_price" binding:"min=0"`
}

type OrderRequest struct {
	UserID          string             `json:"user_id" binding:"required,uuid"`
	Status          string             `json:"status" binding:"required"`
	DiscountPercent int                `json:"discount_percent" binding:"omitempty,min=0,max=100"`
	Items           []OrderItemRequest `json:"items" binding:"required,min=1,dive"`
}

type OrderPathParam struct {
	ID string `uri:"id" binding:"required,uuid"`
}

type OrderListQuery struct {
	UserID string `form:"user_id" binding:"omitempty,uuid"`
	Page   int    `form:"page" binding:"omitempty,min=1"`
	Limit  int    `form:"limit" binding:"omitempty,min=1,max=100"`
}

type OrderItemResponse struct {
	SKU       string `json:"sku"`
	Name      string `json:"name"`
	Quantity  int    `json:"quantity"`
	UnitPrice int64  `json:"unit_price"`
}

type OrderResponse struct {
	ID        string              `json:"id"`
	UserID    string              `json:"user_id"`
	Status    string              `json:"status"`
	Items     []OrderItemResponse `json:"items"`
	Subtotal  int64               `json:"subtotal"`
	Discount  int64               `json:"discount"`
	Total     int64               `json:"total"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}

func (r OrderItemRequest) ToDomain() order.Item {
	return order.Item{
		SKU:       r.SKU,
		Name:      r.Name,
		Quantity:  r.Quantity,
		UnitPrice: r.UnitPrice,
	}
}

func (r OrderRequest) ToCreateInput() order.CreateInput {
	return order.CreateInput{
		UserID:          r.UserID,
		Status:          r.Status,
		DiscountPercent: r.DiscountPercent,
		Items:           r.ItemsToDomain(),
	}
}

func (r OrderRequest) ToUpdateInput() order.UpdateInput {
	return order.UpdateInput{
		UserID:          r.UserID,
		Status:          r.Status,
		DiscountPercent: r.DiscountPercent,
		Items:           r.ItemsToDomain(),
	}
}

func (r OrderRequest) ItemsToDomain() []order.Item {
	items := make([]order.Item, len(r.Items))
	for i, item := range r.Items {
		items[i] = item.ToDomain()
	}
	return items
}

func NewOrderResponse(o *order.Order) OrderResponse {
	if o == nil {
		return OrderResponse{}
	}

	items := make([]OrderItemResponse, len(o.Items))
	for i, item := range o.Items {
		items[i] = OrderItemResponse{
			SKU:       item.SKU,
			Name:      item.Name,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
		}
	}

	return OrderResponse{
		ID:        o.ID,
		UserID:    o.UserID,
		Status:    o.Status,
		Items:     items,
		Subtotal:  o.Subtotal,
		Discount:  o.Discount,
		Total:     o.Total,
		CreatedAt: o.CreatedAt,
		UpdatedAt: o.UpdatedAt,
	}
}

func NewOrderResponses(values []order.Order) []OrderResponse {
	results := make([]OrderResponse, len(values))
	for i := range values {
		results[i] = NewOrderResponse(&values[i])
	}
	return results
}
