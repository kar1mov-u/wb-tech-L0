package main

import (
	"context"
	"log"
	"net/http"
	"order_service/internal/config"
	"order_service/internal/handler"
	"order_service/internal/infra/postgres"
	"order_service/internal/infra/redis"
	"order_service/internal/ports/adapters/cache"
	"order_service/internal/ports/adapters/storage"
	"order_service/internal/service"
)

func main() {

	cnf := config.LoadConfig()

	pool, err := postgres.New(context.Background(), cnf.Postgres)
	if err != nil {
		panic(err)
	}

	redis, err := redis.New(cnf.Redis)
	if err != nil {
		panic(err)
	}

	orderStorage := storage.NewOrderStoragePostgres(pool)
	orderCache := cache.NewOrderCacheRedis(redis)

	orderService := service.NewOrderService(orderStorage, orderCache)

	orderServiceHandler := handler.NewOrderServiceHandler(orderService)
	httpHandler := orderServiceHandler.SetRoutes()

	srv := http.Server{Handler: httpHandler, Addr: ":8081"}

	log.Println("server is listening")
	srv.ListenAndServe()

}
