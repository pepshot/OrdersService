package cache

import (
	"context"
	"sync"

	"OrdersService/internal/order"
)

type OrderCache struct {
	mu     sync.RWMutex
	orders map[string]order.Order
}

func NewOrderCache() *OrderCache {
	return &OrderCache{
		orders: make(map[string]order.Order),
	}
}

func (c *OrderCache) Set(order order.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.orders[order.OrderUID] = order
}

func (c *OrderCache) Get(orderUID string) (order.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	order, ok := c.orders[orderUID]
	return order, ok
}

func (c *OrderCache) GetAll() []order.Order {
	c.mu.RLock()
	defer c.mu.RUnlock()
	orders := make([]order.Order, 0, len(c.orders))
	for _, order := range c.orders {
		orders = append(orders, order)
	}
	return orders
}

func (c *OrderCache) RestoreFromDB(ctx context.Context, getAllOrdersFunc func(ctx context.Context) ([]order.Order, error)) error {
	orders, err := getAllOrdersFunc(ctx)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	for _, order := range orders {
		c.orders[order.OrderUID] = order
	}

	return nil
}
