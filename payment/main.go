package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Shopify/sarama"
)

type paymentRequest struct {
	Service       string `json:"service"`
	ServiceType   string `json:"type"`
	OrderId       string `json:"orderId"`
	CustomerId    string `json:"customerId"`
	PaymentMethod int    `json:"paymentMethod"`
	Price         int    `json:"price"`
}

type paymentResult struct {
	Service     string `json:"service"`
	ServiceType string `json:"type"`
	Result      bool   `json:"result"`
	OrderId     string `json:"orderId"`
	CustomerId  string `json:"customerId"`
}

func main() {

	queue_host := os.Getenv("QUEUE_HOST")
	queue_port := os.Getenv("QUEUE_PORT")
	consume_name := os.Getenv("PAYMENT_NAME")
	produce_name := os.Getenv("PRODUCE_NAME")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals,
		syscall.SIGKILL,
		syscall.SIGTERM,
		syscall.SIGINT,
		os.Interrupt)

	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	brokers := []string{queue_host + ":" + queue_port}

	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := consumer.Close(); err != nil {
			panic(err)
		}
	}()

	partition, err := consumer.ConsumePartition(consume_name, 0, sarama.OffsetNewest)
	if err != nil {
		panic(err)
	}

	configConsume := sarama.NewConfig()
	configConsume.Producer.Return.Errors = true
	configConsume.Producer.Return.Successes = true
	configConsume.Producer.Retry.Max = 3

	producer, err := sarama.NewSyncProducer(brokers, configConsume)
	if err != nil {
		panic(err)
	}
	defer producer.Close()

	go func() {
	CONSUMER_FOR:
		for {
			select {
			case msg := <-partition.Messages():
				var consumed paymentRequest
				if err := json.Unmarshal(msg.Value, &consumed); err != nil {
					fmt.Println(err)
				}
				fmt.Printf("Consumed message : %#v\n", consumed)
				var app Application
				result := app.approvePayment(&consumed)
				b, err := json.Marshal(result)
				if err != nil {
					panic(err)
				}
				conmsg := &sarama.ProducerMessage{
					Topic: produce_name,
					Key:   sarama.StringEncoder(consumed.OrderId),
					Value: sarama.StringEncoder(string(b)),
				}
				partition, offset, err := producer.SendMessage(conmsg)
				if err != nil {
					fmt.Printf("FAILED to send message: %s value: %#v\n", err, result)
				} else {
					fmt.Printf("> message sent to partition %d at offset %d msg %#v\n", partition, offset, result)
				}
			case <-ctx.Done():
				break CONSUMER_FOR
			}
		}
	}()

	fmt.Println("Payment Service start..")

	<-signals

	fmt.Println("Payment Service stop..")
}

type Service interface {
	approvePayment(req *paymentRequest)
}

type Application struct {
}

func (app Application) approvePayment(req *paymentRequest) *paymentResult {

	res := paymentResult{
		Service:     "payment",
		ServiceType: "request",
		OrderId:     req.OrderId,
		CustomerId:  req.CustomerId,
		Result:      true,
	}
	if req.PaymentMethod != 1 {
		res.Result = false
	}

	return &res
}
