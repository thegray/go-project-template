package rest

import (
	"project-template/internal/order"
	"project-template/internal/usecase/checkout"
)

type CheckoutRequest struct {
	UserID string             `json:"user_id" binding:"required,uuid"`
	Items  []OrderItemRequest `json:"items" binding:"required,min=1,dive"`
}

type CheckoutResponse struct {
	User          UserResponse  `json:"user"`
	Order         OrderResponse `json:"order"`
	DiscountGiven int           `json:"discount_given"`
}

func (r CheckoutRequest) ToInput() checkout.Input {
	items := make([]order.Item, len(r.Items))
	for i, item := range r.Items {
		items[i] = item.ToDomain()
	}

	return checkout.Input{
		UserID: r.UserID,
		Items:  items,
	}
}

func NewCheckoutResponse(result *checkout.Result) CheckoutResponse {
	if result == nil {
		return CheckoutResponse{}
	}

	return CheckoutResponse{
		User:          NewUserResponse(result.User),
		Order:         NewOrderResponse(result.Order),
		DiscountGiven: result.DiscountGiven,
	}
}
