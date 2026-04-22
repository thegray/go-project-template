package repository

import (
	"context"
	"errors"

	"project-template/internal/order"
	applogger "project-template/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Repository struct {
	db     *gorm.DB
	logger *applogger.Logger
}

func New(db *gorm.DB, logger *applogger.Logger) *Repository {
	if logger == nil {
		logger = applogger.Wrap(zap.NewNop())
	}
	return &Repository{db: db, logger: logger}
}

func (r *Repository) Create(ctx context.Context, value *order.Order) error {
	log := r.logger
	log.InfoCtx(ctx, "persist order", zap.String("order_id", value.ID), zap.String("user_id", value.UserID))

	model, err := fromDomain(value)
	if err != nil {
		log.ErrorCtx(ctx, "persist order failed: map model", zap.String("order_id", value.ID), zap.Error(err))
		return err
	}

	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		log.ErrorCtx(ctx, "persist order failed", zap.String("order_id", value.ID), zap.Error(err))
		return err
	}
	log.InfoCtx(ctx, "persist order succeeded", zap.String("order_id", value.ID))
	return nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*order.Order, error) {
	log := r.logger
	log.InfoCtx(ctx, "load order by id", zap.String("order_id", id))

	var model pgOrder
	err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.WarnCtx(ctx, "load order by id not found", zap.String("order_id", id))
			return nil, order.ErrNotFound
		}
		log.ErrorCtx(ctx, "load order by id failed", zap.String("order_id", id), zap.Error(err))
		return nil, err
	}

	log.InfoCtx(ctx, "load order by id succeeded", zap.String("order_id", id))
	return model.toDomain()
}

func (r *Repository) ListByUser(ctx context.Context, userID string, page, limit int) ([]order.Order, error) {
	log := r.logger
	log.InfoCtx(ctx, "list orders repository", zap.String("user_id", userID), zap.Int("page", page), zap.Int("limit", limit))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	query := r.db.WithContext(ctx).Model(&pgOrder{})
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	var rows []pgOrder
	if err := query.Order("created_at desc").Limit(limit).Offset((page - 1) * limit).Find(&rows).Error; err != nil {
		log.ErrorCtx(ctx, "list orders repository failed", zap.String("user_id", userID), zap.Error(err))
		return nil, err
	}

	results := make([]order.Order, 0, len(rows))
	for _, row := range rows {
		domain, err := row.toDomain()
		if err != nil {
			return nil, err
		}
		results = append(results, *domain)
	}

	log.InfoCtx(ctx, "list orders repository succeeded", zap.String("user_id", userID), zap.Int("count", len(results)))
	return results, nil
}

func (r *Repository) Update(ctx context.Context, value *order.Order) error {
	log := r.logger
	log.InfoCtx(ctx, "persist order update", zap.String("order_id", value.ID))

	model, err := fromDomain(value)
	if err != nil {
		log.ErrorCtx(ctx, "persist order update failed: map model", zap.String("order_id", value.ID), zap.Error(err))
		return err
	}

	if err := r.db.WithContext(ctx).Save(&model).Error; err != nil {
		log.ErrorCtx(ctx, "persist order update failed", zap.String("order_id", value.ID), zap.Error(err))
		return err
	}
	log.InfoCtx(ctx, "persist order update succeeded", zap.String("order_id", value.ID))
	return nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	log := r.logger
	log.InfoCtx(ctx, "delete order repository", zap.String("order_id", id))

	result := r.db.WithContext(ctx).Delete(&pgOrder{}, "id = ?", id)
	if result.Error != nil {
		log.ErrorCtx(ctx, "delete order repository failed", zap.String("order_id", id), zap.Error(result.Error))
		return result.Error
	}
	if result.RowsAffected == 0 {
		log.WarnCtx(ctx, "delete order repository not found", zap.String("order_id", id))
		return order.ErrNotFound
	}
	log.InfoCtx(ctx, "delete order repository succeeded", zap.String("order_id", id))
	return nil
}
