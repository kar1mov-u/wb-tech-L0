package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"order_service/internal/config"
	"order_service/internal/handler"
	"order_service/internal/infra/kafka"
	"order_service/internal/infra/postgres"
	"order_service/internal/infra/redis"
	"order_service/internal/models"
	"order_service/internal/ports/adapters/cache"
	"order_service/internal/ports/adapters/reciever"
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

	kafkaReader := kafka.NewReader(cnf.Kafka)
	err = kafka.CreateTopicIfNotExists(cnf.Kafka)
	if err != nil {
		panic(err)
	}

	kafkaReciever := reciever.NewRecieverKafka(kafkaReader, func(b []byte) (models.Order, error) {
		var o models.Order
		err := json.Unmarshal(b, &o)
		return o, err
	})
	orderRecieverService := service.NewOrderRecieverService(kafkaReciever, orderService.SaveOrder)

	go func() {
		log.Printf("reciver is listenign on port : %s, topic:%s", cnf.Kafka.Host, cnf.Kafka.Topic)
		err := orderRecieverService.Run(context.TODO())
		if err != nil {
			panic(err)
		}
	}()

	httpHandler := orderServiceHandler.SetRoutes()

	srv := http.Server{Handler: httpHandler, Addr: ":8081"}

	log.Println("server is listening")
	srv.ListenAndServe()

}
