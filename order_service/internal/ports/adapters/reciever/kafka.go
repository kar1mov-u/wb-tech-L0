package reciever

import (
	"context"
	"errors"
	"log"

	"github.com/segmentio/kafka-go"
)

type ReceiverKafka[M any] struct {
	kafkaReader *kafka.Reader
	decodeFn    func([]byte) (M, error)
}

func NewRecieverKafka[M any](r *kafka.Reader, f func([]byte) (M, error)) *ReceiverKafka[M] {
	return &ReceiverKafka[M]{kafkaReader: r, decodeFn: f}
}

func (r *ReceiverKafka[M]) Run(ctx context.Context, handle func(context.Context, M) error) error {

	//will recieve event
	for {
		//read the message
		msg, err := r.kafkaReader.ReadMessage(ctx)
		log.Println("read the message")
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return err
		}

		//decode payload
		m, err := r.decodeFn(msg.Value)
		if err != nil {
			log.Printf("falied to decode kafka message: %v", err)
			continue
		}
		//to-do try to valdiate

		//process the order
		err = handle(ctx, m)
		if err != nil {
			log.Printf("failed to process the kafka order: %v", err)
			continue
		}
	}

}
