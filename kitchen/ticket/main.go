package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Shopify/sarama"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	graphql "github.com/hasura/go-graphql-client"
	"github.com/jinzhu/gorm"
)

type ConsumedMessage struct {
	Service     string `json:"service"`
	ServiceType string `json:"type"`
	OrderId     string `json:"orderId"`
	CustomerId  string `json:"customerId"`
	Items       []Item `json:"items"`
}

type Item struct {
	MenuId   string `json:"menuId"`
	Quantity int    `json:"quantity"`
}

type CancelMessage struct {
	Service     string `json:"service"`
	ServiceType string `json:"type"`
	OrderId     string `json:"orderId"`
}

type ProduceMessage struct {
	Service     string `json:"service"`
	ServiceType string `json:"type"`
	OrderId     string `json:"orderId"`
	Result      bool   `json:"result"`
	Price       int    `json:"price"`
}

func main() {

	db_user := os.Getenv("DB_USER")
	db_pass := os.Getenv("DB_PASS")
	db_host := os.Getenv("DB_HOST")
	db_port := os.Getenv("DB_PORT")
	db_name := os.Getenv("DB_NAME")
	db_svc := os.Getenv("DB_SERVICE")
	queue_host := os.Getenv("QUEUE_HOST")
	queue_port := os.Getenv("QUEUE_PORT")
	consume_name := os.Getenv("CONSUME_NAME")
	cancel_name := os.Getenv("CANCEL_CONSUME_NAME")
	produce_name := os.Getenv("PRODUCE_NAME")

	dataSource := db_user + ":" + db_pass + "@tcp(" + db_host + ":" + db_port + ")/" + db_name + "?charset=utf8&parseTime=True&loc=Local"

	db, err := gorm.Open(db_svc, dataSource)
	if err != nil {
		panic(err)
	}
	if db == nil {
		panic(err)
	}
	defer func() {
		if db != nil {
			if err := db.Close(); err != nil {
				panic(err)
			}
		}
	}()
	db.LogMode(true)

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

	cancelPartition, err := consumer.ConsumePartition(cancel_name, 0, sarama.OffsetNewest)
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
				var consumed ConsumedMessage
				if err := json.Unmarshal(msg.Value, &consumed); err != nil {
					fmt.Println(err)
				}
				fmt.Printf("Consumed message : %#v\n", consumed)
				result, err := ticketAppService(consumed, db)
				b, err := json.Marshal(result)
				if err != nil {
					fmt.Println(err)
				}
				conmsg := &sarama.ProducerMessage{
					Topic: produce_name,
					Key:   sarama.StringEncoder(consumed.OrderId),
					Value: sarama.StringEncoder(string(b)),
				}
				partition, offset, err := producer.SendMessage(conmsg)
				if err != nil {
					log.Printf("FAILED to send message: %s value: %#v\n", err, result)
				} else {
					log.Printf("> message sent to partition %d at offset %d msg %#v\n", partition, offset, result)
				}
			case cmsg := <-cancelPartition.Messages():
				var cancel CancelMessage
				if err := json.Unmarshal(cmsg.Value, &cancel); err != nil {
					fmt.Println(err)
				}
				fmt.Printf("Consumed message : %#v\n", cancel)
				cancelAppService(cancel.OrderId, db)
			case <-ctx.Done():
				break CONSUMER_FOR
			}
		}
	}()

	fmt.Println("Ticket Service start..")

	<-signals

	fmt.Println("Ticket Service stop..")
}

func ticketAppService(consumed ConsumedMessage, db *gorm.DB) (*ProduceMessage, error) {

	message := ProduceMessage{
		Service:     "kitchen",
		ServiceType: "create",
		OrderId:     consumed.OrderId,
	}

	c := &CmdQueryMenu{
		Items: consumed.Items,
	}
	menuLists, err := c.getMenu()
	if err != nil {
		message.Result = false
		return &message, err
	}

	cmd := &CmdTicket{
		OrderId:    consumed.OrderId,
		CustomerId: consumed.CustomerId,
		MenuLists:  menuLists,
	}

	totalPrice, err := issueTicket(cmd, db)
	if err != nil {
		message.Result = false
		return &message, err
	}
	message.Result = true
	message.Price = totalPrice

	return &message, nil

}

type CmdQueryMenu struct {
	Items []Item
}

type MenuList struct {
	MenuId   string
	Quantity int
	Price    int
}

type getMenuInfo interface {
	getMenu() []*MenuList
}

func (c *CmdQueryMenu) getMenu() ([]*MenuList, error) {

	qraghqlUrl := os.Getenv("GRAPHQL_URL")

	client := graphql.NewClient(qraghqlUrl, nil)

	var results []*MenuList

	for _, item := range c.Items {

		var q struct {
			Menu struct {
				Id    graphql.ID
				Price graphql.Int
			} `graphql:"menu(id: $id)"`
		}

		variables := map[string]interface{}{
			"id": graphql.ID(item.MenuId),
		}

		err := client.Query(context.Background(), &q, variables)
		if err != nil {
			fmt.Printf("Failed get menu %#v¥n", item)
			return nil, err
		}

		results = append(results, &MenuList{
			MenuId:   item.MenuId,
			Quantity: item.Quantity,
			Price:    int(q.Menu.Price),
		})
	}
	if len(results) < 1 {
		fmt.Println("Failed get menus")
	}
	return results, nil
}

type CmdTicket struct {
	OrderId    string
	CustomerId string
	MenuLists  []*MenuList
}

type DB struct {
	db *gorm.DB
}

type Repository interface {
	InsertTicket(t *Ticket) error
	CancelTicket(orderId string) error
}

type Ticket struct {
	ID          string `gorm:"column:ticket_id;primary_key"`
	Menu_id     string `gorm:"column:menu_id"`
	Quantity    int    `gorm:"column:quantity"`
	Price       int    `gorm:"column:price"`
	Order_id    string `gorm:"column:order_id"`
	Customer_id string `gorm:"column:customer_id"`
	Status      string `gorm:"column:status"`
}

func (t *Ticket) TableName() string {
	return "ticket"
}

func issueTicket(cmd *CmdTicket, db *gorm.DB) (int, error) {

	var totalPrice int
	for _, menu := range cmd.MenuLists {
		id := GenUUID()
		t := &Ticket{
			ID:          id,
			Menu_id:     menu.MenuId,
			Quantity:    menu.Quantity,
			Price:       menu.Price,
			Order_id:    cmd.OrderId,
			Customer_id: cmd.CustomerId,
			Status:      "APPROVE",
		}
		d := DB{db: db}
		err := d.InsertTicket(t)
		if err != nil {
			return 0, err
		}
		totalPrice += menu.Price * menu.Quantity
	}
	return totalPrice, nil
}

func (d *DB) InsertTicket(ticket *Ticket) error {
	res := d.db.Create(ticket)
	if err := res.Error; err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (d *DB) CancelTicket(orderId string) error {
	res := d.db.Model(&Ticket{}).Where("order_id = ?", orderId).Update("status", "CANCEL")
	if err := res.Error; err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func GenUUID() string {
	u, err := uuid.NewRandom()
	if err != nil {
		fmt.Println(err)
	}
	return u.String()
}

func cancelAppService(orderId string, db *gorm.DB) error {
	d := DB{db: db}
	err := d.CancelTicket(orderId)
	if err != nil {
		return err
	}
	return nil
}
