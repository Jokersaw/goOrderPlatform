package tests

import (
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/jokersaw/goOrderPlatform/internal/cache"
	"github.com/jokersaw/goOrderPlatform/internal/db"
)

func TestCache(t *testing.T) {
	database, err := sqlx.Open("postgres", "postgres://postgres:qwerty1241@localhost:5432/wbOrders?sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to connect to DB: %v", err)
	}
	defer database.Close()

	orderUID := "c784dfg8h9i0j1test"

	startDB := time.Now()
	orderFromDB, err := db.GetOrder(database, orderUID)
	if err != nil {
		t.Fatalf("Failed to get order from DB: %v", err)
	}
	dbDuration := time.Since(startDB)
	t.Logf("DB fetch duration: %v", dbDuration)

	c := cache.NewOrderCache()
	c.Set(orderFromDB)

	startCache := time.Now()
	_, ok := c.Get(orderUID)
	if !ok {
		t.Fatalf("Order not found in cache")
	}
	cacheDuration := time.Since(startCache)
	t.Logf("Cache fetch duration: %v", cacheDuration)

	if cacheDuration >= dbDuration {
		t.Errorf("Cache read should be faster than DB read")
	}
}
