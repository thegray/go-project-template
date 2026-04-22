package order

import (
	"context"
	"errors"
	"strings"
	"time"

	"project-template/internal/shared"

	applogger "project-template/pkg/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	ErrNotFound       = errors.New("order not found")
	ErrInvalidOrder   = errors.New("invalid order")
	ErrInvalidStatus  = errors.New("status is required")
	ErrInvalidItem    = errors.New("order items must not be empty")
	ErrInvalidPricing = errors.New("order pricing is invalid")
)

type Service struct {
	repo Repository
	log  *applogger.Logger
}

func NewService(repo Repository, logger *applogger.Logger) *Service {
	if logger == nil {
		logger = applogger.Wrap(zap.NewNop())
	}
	return &Service{repo: repo, log: logger}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*Order, error) {
	log := s.log
	log.InfoCtx(ctx, "create order request", zap.String("user_id", input.UserID), zap.Int("item_count", len(input.Items)), zap.Int("discount_percent", input.DiscountPercent))

	order := &Order{
		ID:        newID(),
		UserID:    strings.TrimSpace(input.UserID),
		Status:    normalizeStatus(input.Status),
		Items:     cloneItems(input.Items),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := applyPricing(order, input.DiscountPercent); err != nil {
		return nil, err
	}

	if order.UserID == "" {
		return nil, ErrInvalidOrder
	}

	if err := s.repo.Create(ctx, order); err != nil {
		log.ErrorCtx(ctx, "create order failed", zap.String("order_id", order.ID), zap.Error(err))
		return nil, err
	}
	log.InfoCtx(ctx, "create order succeeded", zap.String("order_id", order.ID), zap.Int64("total", order.Total))
	return order, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*Order, error) {
	log := s.log
	log.InfoCtx(ctx, "fetch order", zap.String("order_id", id))
	result, err := s.repo.GetByID(ctx, id)
	if err != nil {
		log.WarnCtx(ctx, "fetch order failed", zap.String("order_id", id), zap.Error(err))
		return nil, err
	}
	log.InfoCtx(ctx, "fetch order succeeded", zap.String("order_id", id))
	return result, nil
}

func (s *Service) List(ctx context.Context, userID string, pagination shared.Pagination) ([]Order, error) {
	log := s.log
	log.InfoCtx(ctx, "list orders", zap.String("user_id", userID), zap.Int("page", pagination.Page), zap.Int("limit", pagination.Limit))
	pagination = pagination.Normalized()
	result, err := s.repo.ListByUser(ctx, userID, pagination.Page, pagination.Limit)
	if err != nil {
		log.WarnCtx(ctx, "list orders failed", zap.String("user_id", userID), zap.Error(err))
		return nil, err
	}
	log.InfoCtx(ctx, "list orders succeeded", zap.String("user_id", userID), zap.Int("count", len(result)))
	return result, nil
}

func (s *Service) Update(ctx context.Context, id string, input UpdateInput) (*Order, error) {
	log := s.log
	log.InfoCtx(ctx, "update order request", zap.String("order_id", id), zap.String("user_id", input.UserID), zap.Int("item_count", len(input.Items)), zap.Int("discount_percent", input.DiscountPercent))

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		log.WarnCtx(ctx, "update order failed: fetch existing", zap.String("order_id", id), zap.Error(err))
		return nil, err
	}

	if strings.TrimSpace(input.UserID) != "" {
		existing.UserID = strings.TrimSpace(input.UserID)
	}
	if strings.TrimSpace(input.Status) != "" {
		existing.Status = normalizeStatus(input.Status)
	}
	if len(input.Items) > 0 {
		existing.Items = cloneItems(input.Items)
	}

	existing.UpdatedAt = time.Now().UTC()
	needsReprice := len(input.Items) > 0 || input.DiscountPercent > 0
	if needsReprice {
		if err := applyPricing(existing, input.DiscountPercent); err != nil {
			return nil, err
		}
	}

	if err := s.repo.Update(ctx, existing); err != nil {
		log.ErrorCtx(ctx, "update order failed", zap.String("order_id", id), zap.Error(err))
		return nil, err
	}
	log.InfoCtx(ctx, "update order succeeded", zap.String("order_id", id), zap.Int64("total", existing.Total))
	return existing, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	log := s.log
	log.InfoCtx(ctx, "delete order request", zap.String("order_id", id))
	if err := s.repo.Delete(ctx, id); err != nil {
		log.WarnCtx(ctx, "delete order failed", zap.String("order_id", id), zap.Error(err))
		return err
	}
	log.InfoCtx(ctx, "delete order succeeded", zap.String("order_id", id))
	return nil
}

func applyPricing(order *Order, discountPercent int) error {
	if len(order.Items) == 0 {
		return ErrInvalidItem
	}
	if discountPercent < 0 || discountPercent > 100 {
		return ErrInvalidPricing
	}

	var subtotal int64
	for _, item := range order.Items {
		if strings.TrimSpace(item.Name) == "" || strings.TrimSpace(item.SKU) == "" || item.Quantity <= 0 || item.UnitPrice < 0 {
			return ErrInvalidItem
		}
		subtotal += int64(item.Quantity) * item.UnitPrice
	}

	discount := subtotal * int64(discountPercent) / 100
	order.Subtotal = subtotal
	order.Discount = discount
	order.Total = subtotal - discount
	return nil
}

func normalizeStatus(status string) string {
	if strings.TrimSpace(status) == "" {
		return "pending"
	}
	return strings.ToLower(strings.TrimSpace(status))
}

func cloneItems(items []Item) []Item {
	cloned := make([]Item, len(items))
	copy(cloned, items)
	return cloned
}

func newID() string {
	return uuid.NewString()
}
