package ports

import (
	"context"
	"order_service/internal/models"
)

type OrderStorage interface {
	GetOrderByID(ctx context.Context, id string) (models.Order, error)
	// GetLastOrders(ctx context.Context, limit int) ([]models.Order, error)
	SaveOrder(ctx context.Context, order models.Order) error
}

type OrderCache interface {
	Set(ctx context.Context, id string, order models.Order) error
	Get(ctx context.Context, id string) (models.Order, bool, error)
	// GetKeys() []string
	// GetKeysAmount() int
}

type Reciever interface {
	Consume(ctx context.Context)
}
