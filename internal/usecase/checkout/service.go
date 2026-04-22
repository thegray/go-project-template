package checkout

import (
	"context"
	"errors"
	"strings"

	"project-template/internal/order"
	"project-template/internal/user"

	applogger "project-template/pkg/logger"

	"go.uber.org/zap"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrInvalidCheckout = errors.New("invalid checkout request")
)

type Service struct {
	users  UserProvider
	orders OrderCreator
	log    *applogger.Logger
}

func NewService(users UserProvider, orders OrderCreator, logger *applogger.Logger) *Service {
	if logger == nil {
		logger = applogger.Wrap(zap.NewNop())
	}
	return &Service{users: users, orders: orders, log: logger}
}

func (s *Service) Checkout(ctx context.Context, input Input) (*Result, error) {
	log := s.log
	log.InfoCtx(ctx, "checkout request", zap.String("user_id", input.UserID), zap.Int("item_count", len(input.Items)))

	if strings.TrimSpace(input.UserID) == "" || len(input.Items) == 0 {
		log.WarnCtx(ctx, "invalid checkout request", zap.String("user_id", input.UserID))
		return nil, ErrInvalidCheckout
	}

	account, err := s.users.GetByID(ctx, input.UserID)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			log.WarnCtx(ctx, "checkout user not found", zap.String("user_id", input.UserID))
			return nil, ErrUserNotFound
		}
		log.ErrorCtx(ctx, "checkout user lookup failed", zap.String("user_id", input.UserID), zap.Error(err))
		return nil, err
	}

	discount := 0
	if strings.Contains(strings.ToLower(account.Email), "vip") {
		discount = 10
	}
	log.InfoCtx(ctx, "checkout pricing resolved", zap.String("user_id", input.UserID), zap.Int("discount_percent", discount))

	createdOrder, err := s.orders.Create(ctx, order.CreateInput{
		UserID:          input.UserID,
		Status:          "checkout_completed",
		DiscountPercent: discount,
		Items:           input.Items,
	})
	if err != nil {
		log.ErrorCtx(ctx, "checkout order creation failed", zap.String("user_id", input.UserID), zap.Error(err))
		return nil, err
	}

	log.InfoCtx(ctx, "checkout succeeded", zap.String("user_id", input.UserID), zap.String("order_id", createdOrder.ID))
	return &Result{
		User:          account,
		Order:         createdOrder,
		DiscountGiven: discount,
	}, nil
}
