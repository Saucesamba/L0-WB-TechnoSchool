package main

import (
	"context"
	"encoding/json"
	"github.com/go-faker/faker/v4"
	"github.com/segmentio/kafka-go"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"
)

type Order struct {
	OrderUID          string `faker:"uuid_digit" json:"order_uid"`
	TrackNumber       string `faker:"len=13" json:"track_number"`
	Entry             string `faker:"oneof:WBIL,WBOL,WBUL,WBAL,WBEL" json:"entry"`
	Locale            string `faker:"oneof:en,ru,kz,by" json:"locale"`
	InternalSignature string `faker:"oneof:-,identifier" json:"internal_signature"`
	CustomerID        string `faker:"username" json:"customer_id"`
	DeliveryService   string `faker:"oneof:Почта России,BoxBerry,CDEK,5POST,Avito" json:"delivery_service"`
	Shardkey          string `faker:"oneof:1,2,3,4,5" json:"shardkey"`
	SmID              int    `faker:"oneof:1,52,4,23,76" json:"sm_id"`
	DateCreated       string `faker:"timestamp" json:"date_created"`
	OofShard          string `faker:"oneof:1,2,3,4,5" json:"oof_shard"`

	Delivery Delivery `json:"delivery"`
	Payment  Payment  `json:"payment"`
	Items    []Item   `json:"items"`
}

type Delivery struct {
	Name    string `faker:"name" json:"name"`
	Phone   string `faker:"phone_number" json:"phone"`
	Zip     string `faker:"word" json:"zip"`
	City    string `faker:"word" json:"city"`
	Address string `faker:"sentence len=3" json:"address"`
	Region  string `faker:"word" json:"region"`
	Email   string `faker:"email" json:"email"`
}

type fakedig struct {
	Num1 float64 `faker:"amount"`
	Num2 float64 `faker:"amount"`
	Num3 float64 `faker:"amount"`
}

type Payment struct {
	TransactionNumber string `faker:"uuid_digit" json:"transaction"`
	RequestID         string `faker:"oneof:-" json:"request_id"`
	Currency          string `faker:"currency" json:"currency"`
	Provider          string `faker:"oneof:wbpay,alfabank,sber,tbank,vtb" json:"provider"`
	Amount            int    `json:"amount"`
	PaymentDT         int64  `faker:"unix_time" json:"payment_dt"`
	Bank              string `faker:"cc_type" json:"bank"`
	DeliveryCost      int    `json:"delivery_cost"`
	GoodsTotal        int    `faker:"oneof:1,2,3,4,5,6,7,8,9,10" json:"goods_total"`
	CustomFee         int    `json:"custom_fee"`
}

type Item struct {
	ChrtID      int    `json:"chrt_id"`
	TrackNumber string `faker:"len=13" json:"track_number"`
	Price       int    `json:"price"`
	RID         string `faker:"len=19" json:"rid"`
	ItemName    string `faker:"word" json:"name"`
	Sale        int    `faker:"oneof:10,20,30,40,50" json:"sale"`
	ItemSize    string `faker:"len=2" json:"size"`
	TotalPrice  int    `json:"total_price"`
	NmID        int    `json:"nm_id"`
	Brand       string `faker:"word" json:"brand"`
	Status      int    `faker:"oneof:200,400,404" json:"status"`
}

const (
	numOrders     = 152
	numGoroutines = 10
)

func AdrToArr(adr string) []string {
	arr := strings.Split(adr, ":")
	arr[1] = strings.TrimSuffix(arr[1], " City")
	arr[2] = strings.TrimSuffix(arr[2], " State")
	arr[3] = strings.TrimSuffix(arr[3], " PostalCode")
	arr[4] = strings.TrimSuffix(arr[4], " Coordinates")
	return arr
}
func main() {
	brokerAddress := "localhost:29092"
	topic := "orders"

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{brokerAddress},
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	})
	defer writer.Close()

	orders := make([]Order, numOrders)
	for i := 0; i < numOrders; i++ {
		err := faker.FakeData(&orders[i])
		orders[i].DateCreated = time.Now().Format("2006-01-02T15:04:05Z07:00")
		strings.ToUpper(orders[i].TrackNumber)

		if err != nil {
			log.Fatalf("Failed to generate fake order data: %v", err)
		}
		orders[i].Delivery = Delivery{}
		faker.FakeData(&orders[i].Delivery)
		am := fakedig{}
		faker.FakeData(&am)

		orders[i].Payment = Payment{}
		faker.FakeData(&orders[i].Payment)
		orders[i].Payment.Amount = int(am.Num1)
		orders[i].Payment.DeliveryCost = int(am.Num2)
		orders[i].Payment.CustomFee = int(am.Num3)
		orders[i].Items = []Item{}
		for j := 0; j < 3; j++ {
			var item Item
			faker.FakeData(&item)
			item.ChrtID = rand.Intn(10000)
			item.Price = rand.Intn(100000)
			item.TotalPrice = item.Price * (100 - item.Sale) / 100
			item.NmID = rand.Intn(1000)
			orders[i].Items = append(orders[i].Items, item)
		}
	}

	chunkSize := numOrders / numGoroutines
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if i == numGoroutines-1 {
			end = numOrders
		}

		wg.Add(1)
		go func(orders []Order) {
			defer wg.Done()
			for _, order := range orders {
				orderBytes, err := json.Marshal(order)
				if err != nil {
					log.Printf("Failed to marshal order: %v", err)
					continue
				}

				msg := kafka.Message{
					Key:   []byte(order.OrderUID),
					Value: orderBytes,
				}

				err = writer.WriteMessages(context.Background(), msg)
				if err != nil {
					log.Printf("Failed to write messages: %v", err)
					continue
				}

				log.Printf("Sent order: %s", order.OrderUID)
				time.Sleep(time.Second * 4)
			}
		}(orders[start:end])
	}

	wg.Wait()

	log.Println("Producer finished")
}
