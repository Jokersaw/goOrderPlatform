package main

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/jokersaw/goOrderPlatform/internal/cache"
	"github.com/jokersaw/goOrderPlatform/internal/db"
	myKafka "github.com/jokersaw/goOrderPlatform/internal/kafka"
	webserver "github.com/jokersaw/goOrderPlatform/internal/webServer"
)

var (
	orderCache *cache.OrderCache
	database   *sqlx.DB
)

func main() {

	database = db.EstablishConnection()

	defer database.Close()

	orderCache = cache.NewOrderCache()

	orders, err := db.GetAllOrders(database)
	if err != nil {
		log.Fatalf("failed to preload orders: %v", err)
	}
	orderCache.SetAll(orders)
	log.Printf("Cache was restored (%d orders)", len(orders))

	orderChan, err := myKafka.ConsumeOrders([]string{"localhost:9092"}, "orders")
	if err != nil {
		log.Fatalf("Failed to start Kafka consumer: %v", err)
	}

	go func() {
		for order := range orderChan {
			err := db.InsertOrder(database, order)
			if err != nil {
				log.Printf("Failed to insert order %s: %v", order.OrderUID, err)
			}
		}
	}()

	srv := webserver.NewServer(database, orderCache, "./frontend")
	if err := srv.Start(":8080"); err != nil {
		log.Fatalf("HTTP server error: %v", err)
	}
}
