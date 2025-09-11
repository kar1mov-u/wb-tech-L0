package service

import (
	"context"
	"order_service/internal/models"
	"order_service/internal/ports/adapters/reciever"
)

type OrderReciverService struct {
	reciever         *reciever.ReceiverKafka[models.Order]
	orderProcessFunc func(ctx context.Context, order models.Order) error
}

func NewOrderRecieverService(reciever *reciever.ReceiverKafka[models.Order], f func(ctx context.Context, order models.Order) error) *OrderReciverService {
	return &OrderReciverService{
		reciever:         reciever,
		orderProcessFunc: f,
	}
}

func (o *OrderReciverService) Run(ctx context.Context) error {
	return o.reciever.Run(ctx, o.orderProcessFunc)
}
