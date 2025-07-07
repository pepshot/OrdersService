package repository

import (
	"OrdersService/internal/order"
	"OrdersService/pkg/client/postgresql"
	"OrdersService/pkg/logging"
	"context"
	"fmt"
	"strings"
)

type IRepository interface {
	CreateOrder(ctx context.Context, order *order.Order) error
	FindAll(ctx context.Context) (order []order.Order, err error)
	FindOne(ctx context.Context, orderUID string) (*order.Order, error)
}

type OrderRepository struct {
	client postgresql.IClient
	logger *logging.Logger
}

func NewOrderRepository(client postgresql.IClient, logger *logging.Logger) *OrderRepository {
	return &OrderRepository{
		client: client,
		logger: logger,
	}
}

func formatQuery(query string) string {
	return strings.ReplaceAll(strings.ReplaceAll(query, "\t", ""), "\n", "")
}

func (r OrderRepository) CreateOrder(ctx context.Context, order *order.Order) error {
	panic("implement me")
}

func (r OrderRepository) FindAll(ctx context.Context) (orders []order.Order, err error) {
	ordersUIDQuery := `SELECT order_uid FROM public.order`
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", formatQuery(ordersUIDQuery)))

	rows, err := r.client.Query(ctx, ordersUIDQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders = make([]order.Order, 0)
	for rows.Next() {
		var orderUID string

		if err := rows.Scan(&orderUID); err != nil {
			return nil, err
		}

		newOrder, err := r.FindOne(ctx, orderUID)

		if err != nil {
			return nil, err
		}

		if newOrder != nil {
			orders = append(orders, *newOrder)
		}
	}

	return orders, nil
}

func (r OrderRepository) FindOne(ctx context.Context, orderUID string) (*order.Order, error) {
	var newOrder order.Order

	// Get order
	orderQuery := `
		SELECT 
			order_uid, track_number, entry, ocale, internal_signature, 
			customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
		FROM public.order 
		WHERE order_uid = $1
	`
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", formatQuery(orderQuery)))

	err := r.client.QueryRow(ctx, orderQuery, orderUID).Scan(
		&newOrder.OrderUID,
		&newOrder.TrackNumber,
		&newOrder.Entry,
		&newOrder.Locale,
		&newOrder.InternalSignature,
		&newOrder.CustomerID,
		&newOrder.DeliveryService,
		&newOrder.Shardkey,
		&newOrder.SmID,
		&newOrder.DateCreated,
		&newOrder.OofShard,
	)

	if err != nil {
		return &order.Order{}, err
	}

	// Get delivery
	var delivery order.Delivery
	deliveryQuery := `
		SELECT 
			name, phone, zip, city, address, region, email
		FROM public.delivery
		WHERE order_uid = $1
	`
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", formatQuery(deliveryQuery)))

	err = r.client.QueryRow(ctx, deliveryQuery, orderUID).Scan(
		&delivery.Name,
		&delivery.Phone,
		&delivery.Zip,
		&delivery.City,
		&delivery.Address,
		&delivery.Region,
		&delivery.Email,
	)

	if err != nil {
		return &order.Order{}, err
	}
	newOrder.Delivery = delivery

	// Get payment
	var payment order.Payment
	paymentQuery := `
		SELECT 
			transaction, request_id, currency, provider, amount, 
			payment_dt, bank, delivery_cost, goods_total, custom_fee
		FROM public.payment
		WHERE order_uid = $1
	`
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", formatQuery(paymentQuery)))

	err = r.client.QueryRow(ctx, paymentQuery, orderUID).Scan(
		&payment.Transaction,
		&payment.RequestID,
		&payment.Currency,
		&payment.Provider,
		&payment.Amount,
		&payment.PaymentDt,
		&payment.Bank,
		&payment.DeliveryCost,
		&payment.GoodsTotal,
		&payment.CustomFee,
	)

	if err != nil {
		return &order.Order{}, err
	}
	newOrder.Payment = payment

	// Get items
	itemsQuery := `
		SELECT 
			chrt_id, track_number, price, rid, name, sale, 
			size, total_price, nm_id, brand, status
		FROM public.items 
		WHERE order_uid = $1
	`
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", formatQuery(itemsQuery)))

	rows, err := r.client.Query(ctx, itemsQuery, orderUID)
	if err != nil {
		return &order.Order{}, err
	}
	defer rows.Close()

	var items []order.Item
	for rows.Next() {
		var item order.Item
		err = rows.Scan(
			&item.ChrtID,
			&item.TrackNumber,
			&item.Price,
			&item.Rid,
			&item.Name,
			&item.Sale,
			&item.Size,
			&item.TotalPrice,
			&item.NmID,
			&item.Brand,
			&item.Status,
		)

		if err != nil {
			return &order.Order{}, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return &order.Order{}, err
	}

	newOrder.Items = items

	return &newOrder, nil
}
