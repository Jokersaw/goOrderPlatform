package cache

import (
	"sync"

	"github.com/jokersaw/goOrderPlatform/internal/models"
)

type OrderCache struct {
	mu    sync.RWMutex
	store map[string]models.OrderMessage
}

func NewOrderCache() *OrderCache {
	return &OrderCache{
		store: make(map[string]models.OrderMessage),
	}
}

func (c *OrderCache) Get(orderUID string) (models.OrderMessage, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	order, ok := c.store[orderUID]
	return order, ok
}

func (c *OrderCache) Set(order models.OrderMessage) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[order.OrderUID] = order
}

func (c *OrderCache) SetAll(orders map[string]models.OrderMessage) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store = orders
}
