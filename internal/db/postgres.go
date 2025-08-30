package db

import (
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/jokersaw/goOrderPlatform/internal/models"
	_ "github.com/lib/pq"
)

func EstablishConnection() *sqlx.DB {
	connStr := "postgres://postgres:qwerty1241@localhost:5432/wbOrders?sslmode=disable"

	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}

	return db
}

func InsertOrder(db *sqlx.DB, order models.OrderMessage) error {
	tx := db.MustBegin()

	_, err := tx.Exec(`
		INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature,
							customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	`, order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature,
		order.CustomerID, order.DeliveryService, order.ShardKey, order.SmID, order.DateCreated, order.OofShard)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO deliveries (order_uid, name, phone, zip, city, address, region, email)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`, order.OrderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip,
		order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	paymentTime := time.Unix(order.Payment.PaymentDT, 0)
	_, err = tx.Exec(`
		INSERT INTO payments (order_uid, transaction_id, request_id, currency, provider, amount,
							  payment_dt, bank, delivery_cost, goods_total, custom_fee)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	`, order.OrderUID, order.Payment.TransactionID, order.Payment.RequestID, order.Payment.Currency,
		order.Payment.Provider, order.Payment.Amount, paymentTime, order.Payment.Bank,
		order.Payment.DeliveryCost, order.Payment.GoodsTotal, order.Payment.CustomFee)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	for _, item := range order.Items {
		_, err = tx.Exec(`
			INSERT INTO items (order_uid, chrt_id, track_number, price, rid, name, sale,
							   size, total_price, nm_id, brand, status)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		`, order.OrderUID, item.ChrtID, item.TrackNumber, item.Price, item.RID, item.Name,
			item.Sale, item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Printf("Inserted Order %s", order.OrderUID)
	return nil
}

func GetOrder(db *sqlx.DB, orderUID string) (models.OrderMessage, error) {
	var orders []models.OrderMessage
	if err := db.Select(&orders, `SELECT * FROM orders WHERE order_uid=$1`, orderUID); err != nil {
		return models.OrderMessage{}, err
	}

	var deliveries []models.Delivery
	if err := db.Select(&deliveries, `SELECT * FROM deliveries WHERE order_uid=$1`, orderUID); err != nil {
		return models.OrderMessage{}, err
	}

	var payments []models.Payment
	if err := db.Select(&payments, `
	SELECT
		id, order_uid, transaction_id, request_id, currency, provider, amount,
		EXTRACT(EPOCH FROM payment_dt)::bigint AS payment_dt,
		bank, delivery_cost, goods_total, custom_fee
	FROM payments
	WHERE order_uid = $1
	`, orderUID); err != nil {
		return models.OrderMessage{}, err
	}

	var items []models.Item
	if err := db.Select(&items, `SELECT * FROM items WHERE order_uid=$1`, orderUID); err != nil {
		return models.OrderMessage{}, err
	}

	result := buildOrders(orders, deliveries, payments, items)

	_, ok := result[orderUID]
	if !ok {
		return models.OrderMessage{}, fmt.Errorf("order %s not found", orderUID)
	}
	return result[orderUID], nil
}

func GetAllOrders(db *sqlx.DB) (map[string]models.OrderMessage, error) {
	var orders []models.OrderMessage
	if err := db.Select(&orders, `SELECT * FROM orders`); err != nil {
		return nil, err
	}

	var deliveries []models.Delivery
	if err := db.Select(&deliveries, `SELECT * FROM deliveries`); err != nil {
		return nil, err
	}

	var payments []models.Payment
	if err := db.Select(&payments, `
	SELECT
		id, order_uid, transaction_id, request_id, currency, provider, amount,
		EXTRACT(EPOCH FROM payment_dt)::bigint AS payment_dt,
		bank, delivery_cost, goods_total, custom_fee
	FROM payments
	`); err != nil {
		return nil, err
	}

	var items []models.Item
	if err := db.Select(&items, `SELECT * FROM items`); err != nil {
		return nil, err
	}

	result := buildOrders(orders, deliveries, payments, items)

	return result, nil
}

func buildOrders(
	orders []models.OrderMessage,
	deliveries []models.Delivery,
	payments []models.Payment,
	items []models.Item,
) map[string]models.OrderMessage {
	result := make(map[string]models.OrderMessage)

	for _, o := range orders {
		result[o.OrderUID] = o
	}

	for _, d := range deliveries {
		if order, ok := result[d.OrderUID]; ok {
			order.Delivery = d
			result[d.OrderUID] = order
		}
	}

	for _, p := range payments {
		if order, ok := result[p.OrderUID]; ok {
			order.Payment = p
			result[p.OrderUID] = order
		}
	}

	for _, it := range items {
		if order, ok := result[it.OrderUID]; ok {
			order.Items = append(order.Items, it)
			result[it.OrderUID] = order
		}
	}

	return result
}
