package service

import (
	"context"
	"log"
	"order_service/internal/models"
	"order_service/internal/ports"
)

type OrderService struct {
	storage ports.OrderStorage
	cache   ports.OrderCache
}

func NewOrderService(storage ports.OrderStorage, cache ports.OrderCache) *OrderService {
	return &OrderService{storage: storage, cache: cache}
}

func (s *OrderService) GetOrder(ctx context.Context, id string) (models.Order, error) {

	var order models.Order

	//first try to get from the cache
	order, ok, err := s.cache.Get(ctx, id)
	if ok {
		return order, nil
	}
	if err != nil {
		log.Printf("error on getting from cache: %v", err)
	}

	//try to get from the storage
	order, err = s.storage.GetOrderByID(ctx, id)
	if err != nil {
		return order, err
	}

	//save to the cache for later use
	go func() {
		err = s.cache.Set(ctx, id, order)
		if err != nil {
			log.Printf("error on setting to cachke :%v", err)
		}
	}()

	return order, nil
}

// func (s *OrderService) FillCache(ctx context.Context, limit int) {
// 	//get orders from the storage
// 	orders, err := s.storage.GetLastOrders(ctx, limit)
// 	if err != nil {
// 		log.Printf("error on getting last orders: %v \n", err)
// 		return
// 	}

// 	//save to the cache
// 	for _, order := range orders {
// 		err = s.cache.Set(ctx, order.OrderUID, order)
// 		if err != nil {
// 			log.Printf("error on setting to cache: %v \n", err)
// 		}
// 	}
// }
