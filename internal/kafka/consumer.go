package kafka

import (
	"OrdersService/internal/cache"
	"OrdersService/internal/config"
	"OrdersService/internal/order"
	"OrdersService/internal/repository"
	"OrdersService/pkg/logging"
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader   *kafka.Reader
	repo     *repository.OrderRepository
	cache    *cache.OrderCache
	logger   *logging.Logger
	stopChan chan struct{}
}

func NewConsumer(cfg *config.Config, repo *repository.OrderRepository, cache *cache.OrderCache, logger *logging.Logger) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Kafka.Brokers,
		GroupID:        cfg.Kafka.GroupID,
		Topic:          cfg.Kafka.Topic,
		MinBytes:       cfg.Kafka.MinBytes,
		MaxBytes:       cfg.Kafka.MaxBytes,
		CommitInterval: time.Second,
	})

	return &Consumer{
		reader:   reader,
		repo:     repo,
		cache:    cache,
		logger:   logger,
		stopChan: make(chan struct{}),
	}
}

func (c *Consumer) Start() {
	c.logger.Trace("Starting Kafka consumer")
	close(c.stopChan)
	if err := c.reader.Close(); err != nil {
		c.logger.Errorf("Error to close Kafka reader: %v", err)
	}
}

func (c *Consumer) Stop() {
	c.logger.Trace("Stopping Kafka consumer")
	close(c.stopChan)
	if err := c.reader.Close(); err != nil {
		c.logger.Errorf("Failed to close Kafka reader: %v", err)
	}
}

func (c *Consumer) consumeMessages() {
	for {
		select {
		case <-c.stopChan:
			return
		default:
			message, err := c.reader.FetchMessage(context.Background())
			if err != nil {
				c.logger.Errorf("Failed to fetch message: %v", err)
				continue
			}

			var order order.Order
			if err = json.Unmarshal(message.Value, &order); err != nil {
				c.logger.Errorf("Failed to unmarshal message: %v", err)
				continue
			}

			if order.OrderUID == "" {
				c.logger.Error("Received order with empty UID")
				continue
			}

			if err := c.processOrder(order); err != nil {
				c.logger.Errorf("Failed to process order %s: %v", order.OrderUID, err)
			}

			if err := c.reader.CommitMessages(context.Background(), message); err != nil {
				c.logger.Errorf("Failed to commit message: %v", err)
			}
		}
	}
}

func (c *Consumer) processOrder(order order.Order) error {
	if err := c.repo.CreateOrder(context.Background(), &order); err != nil {
		return err
	}

	c.cache.Set(order)
	c.logger.Infof("Successfully created order %s", order.OrderUID)
	return nil
}
