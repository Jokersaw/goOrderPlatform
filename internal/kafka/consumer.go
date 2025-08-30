package myKafka

import (
	"context"
	"encoding/json"
	"log"
	"path/filepath"
	"strings"

	"github.com/jokersaw/goOrderPlatform/internal/models"
	"github.com/segmentio/kafka-go"
	"github.com/xeipuuv/gojsonschema"
)

func ConsumeOrders(brokers []string, topic string) (<-chan models.OrderMessage, error) {
	orderChan := make(chan models.OrderMessage)

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		GroupID: "orders-consumer",
		Topic:   topic,
	})

	go func() {
		defer reader.Close()
		defer close(orderChan)

		log.Println("Kafka consumer started...")

		absPath, _ := filepath.Abs("./schemas/order_message.json")
		absPath = strings.ReplaceAll(absPath, "\\", "/")

		schemaLoader := gojsonschema.NewReferenceLoader("file:///" + absPath)

		for {
			msg, err := reader.ReadMessage(context.Background())
			if err != nil {
				log.Printf("could not read message: %v", err)
				continue
			}

			documentLoader := gojsonschema.NewBytesLoader(msg.Value)
			result, err := gojsonschema.Validate(schemaLoader, documentLoader)

			if err != nil {
				log.Printf("Schema validation error: %v", err)
				continue
			}

			if !result.Valid() {
				for _, desc := range result.Errors() {
					log.Printf("Invalid order JSON: %s", desc)
				}
				continue
			}

			var order models.OrderMessage
			if err := json.Unmarshal(msg.Value, &order); err != nil {
				log.Printf("JSON parse error: %v", err)
				continue
			}

			orderChan <- order
		}
	}()

	return orderChan, nil
}
