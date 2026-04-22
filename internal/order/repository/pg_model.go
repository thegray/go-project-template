package repository

import (
	"encoding/json"
	"time"

	"project-template/internal/order"

	"gorm.io/datatypes"
)

type pgOrder struct {
	ID        string         `gorm:"column:id;primaryKey;type:text"`
	UserID    string         `gorm:"column:user_id;index;not null"`
	Status    string         `gorm:"column:status;not null"`
	Items     datatypes.JSON `gorm:"column:items;type:jsonb;not null"`
	Subtotal  int64          `gorm:"column:subtotal;not null"`
	Discount  int64          `gorm:"column:discount;not null"`
	Total     int64          `gorm:"column:total;not null"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime"`
}

func (pgOrder) TableName() string {
	return "orders"
}

func (p pgOrder) toDomain() (*order.Order, error) {
	var items []order.Item
	if len(p.Items) > 0 {
		if err := json.Unmarshal(p.Items, &items); err != nil {
			return nil, err
		}
	}

	return &order.Order{
		ID:        p.ID,
		UserID:    p.UserID,
		Status:    p.Status,
		Items:     items,
		Subtotal:  p.Subtotal,
		Discount:  p.Discount,
		Total:     p.Total,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}, nil
}

func fromDomain(o *order.Order) (pgOrder, error) {
	items, err := json.Marshal(o.Items)
	if err != nil {
		return pgOrder{}, err
	}

	return pgOrder{
		ID:        o.ID,
		UserID:    o.UserID,
		Status:    o.Status,
		Items:     datatypes.JSON(items),
		Subtotal:  o.Subtotal,
		Discount:  o.Discount,
		Total:     o.Total,
		CreatedAt: o.CreatedAt,
		UpdatedAt: o.UpdatedAt,
	}, nil
}
