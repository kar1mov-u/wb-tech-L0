package storage

import (
	"context"
	"errors"
	"fmt"
	"order_service/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderStoragePostgres struct {
	pool *pgxpool.Pool
}

// Queryer is a interface to work with a DB connection, enabling us to use *pool or TX with the same code
type Queryer interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
}

func NewOrderStoragePostgres(pool *pgxpool.Pool) *OrderStoragePostgres {
	return &OrderStoragePostgres{pool: pool}
}

func (s *OrderStoragePostgres) SaveOrder(ctx context.Context, order models.Order) error {
	//create a tx, so if one part fails ,everything should fail
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	defer tx.Rollback(ctx)

	if err != nil {
		return fmt.Errorf("failed to BeginTX: %w", err)
	}
	err = saveOrder(ctx, order, tx)
	if err != nil {
		return fmt.Errorf("failed to save Order element: %w", err)
	}

	err = saveDelivery(ctx, tx, order.Delivery, order.OrderUID)
	if err != nil {
		return fmt.Errorf("failed to save delivery element: %w", err)
	}

	err = saveItems(ctx, tx, order.Items, order.OrderUID)
	if err != nil {
		return fmt.Errorf("failed to save item elements: %w", err)
	}

	err = savePayment(ctx, tx, order.Payment, order.OrderUID)
	if err != nil {
		return fmt.Errorf("failed to save payment element: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to save commit TX:%w", err)
	}

	return nil
}

var ErrNotFound = errors.New("order not found")

func (s *OrderStoragePostgres) GetOrderByID(ctx context.Context, id string) (models.Order, error) {
	// 1) Fetch header (single row)
	const headerSQL = `
        SELECT 
            o.order_uid, o.track_number, o.entry, o.locale, o.internal_signature, 
            o.customer_id, o.delivery_service, o.shardkey, o.sm_id, o.date_created, o.oof_shard,
            d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
            p.transaction, p.request_id, p.currency, p.provider, p.amount, p.payment_dt, p.bank, 
            p.delivery_cost, p.goods_total, p.custom_fee
        FROM orders o
        LEFT JOIN deliveries d ON o.order_uid = d.order_uid
        LEFT JOIN payments   p ON o.order_uid = p.order_uid
        WHERE o.order_uid = $1
    `
	var (
		o models.Order
		d models.Delivery
		p models.Payment
	)
	// Scan, handling NULLs: if any LEFT JOIN columns can be NULL, use sql.NullString/NullInt64 or COALESCE(...) in SQL.
	err := s.pool.QueryRow(ctx, headerSQL, id).Scan(
		&o.OrderUID, &o.TrackNumber, &o.Entry, &o.Locale, &o.InternalSignature,
		&o.CustomerID, &o.DeliveryService, &o.ShardKey, &o.SmID, &o.DateCreated, &o.OofShard,
		&d.Name, &d.Phone, &d.Zip, &d.City, &d.Address, &d.Region, &d.Email,
		&p.Transaction, &p.RequestID, &p.Currency, &p.Provider, &p.Amount, &p.PaymentDT, &p.Bank,
		&p.DeliveryCost, &p.GoodsTotal, &p.CustomFee,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Order{}, ErrNotFound
		}
		return models.Order{}, fmt.Errorf("get order header: %w", err)
	}
	o.Delivery = d
	o.Payment = p

	// 2) Fetch items (0..n rows)
	const itemsSQL = `
        SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
        FROM items
        WHERE order_uid = $1
        ORDER BY chrt_id
    `
	rows, err := s.pool.Query(ctx, itemsSQL, id)
	if err != nil {
		return models.Order{}, fmt.Errorf("get order items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var it models.Item
		if err := rows.Scan(
			&it.ChrtID, &it.TrackNumber, &it.Price, &it.Rid, &it.Name, &it.Sale, &it.Size,
			&it.TotalPrice, &it.NmID, &it.Brand, &it.Status,
		); err != nil {
			return models.Order{}, fmt.Errorf("scan item: %w", err)
		}
		o.Items = append(o.Items, it)
	}
	if err := rows.Err(); err != nil {
		return models.Order{}, fmt.Errorf("items rows: %w", err)
	}

	return o, nil
}

// func (s *OrderStoragePostgres) GetLastOrders(ctx context.Context, limit int) ([]models.Order, error) {

// }

func saveOrder(ctx context.Context, order models.Order, q Queryer) error {
	sql := `INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	_, err := q.Exec(ctx, sql, order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature, order.CustomerID, order.DeliveryService, order.ShardKey, order.SmID, order.DateCreated, order.OofShard)
	if err != nil {
		return err
	}
	return nil
}

func saveDelivery(ctx context.Context, q Queryer, delivery models.Delivery, orderID string) error {
	sql := `
		INSERT INTO deliveries (
			order_uid, name, phone, zip, city, address, region, email
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := q.Exec(ctx, sql,
		orderID,
		delivery.Name,
		delivery.Phone,
		delivery.Zip,
		delivery.City,
		delivery.Address,
		delivery.Region,
		delivery.Email,
	)
	return err
}

func savePayment(ctx context.Context, q Queryer, payment models.Payment, orderID string) error {
	sql := `
		INSERT INTO payments (
			order_uid, transaction, request_id, currency, provider,
			amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := q.Exec(ctx, sql,
		orderID,
		payment.Transaction,
		payment.RequestID,
		payment.Currency,
		payment.Provider,
		payment.Amount,
		payment.PaymentDT,
		payment.Bank,
		payment.DeliveryCost,
		payment.GoodsTotal,
		payment.CustomFee,
	)
	return err
}
func saveItems(ctx context.Context, q Queryer, items []models.Item, orderID string) error {
	sql := `INSERT INTO items (order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	batch := &pgx.Batch{}
	for _, item := range items {
		batch.Queue(sql, orderID, item.ChrtID, item.TrackNumber, item.Price, item.Rid, item.Name, item.Sale, item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status)
	}
	res := q.SendBatch(ctx, batch)
	defer res.Close()

	for i := 0; i < len(items); i++ {
		_, err := res.Exec()
		if err != nil {
			return fmt.Errorf("failed to insert item %d :%w", i, err)
		}
	}
	return nil
}
