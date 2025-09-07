package storage

import (
	"context"
	"fmt"
	"order_service/internal/models"
	"time"

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

func (s *OrderStoragePostgres) GetOrderByID(ctx context.Context, id string) (models.Order, error) {
	sql := `SELECT 
        o.order_uid, o.track_number, o.entry, o.locale, o.internal_signature, o.customer_id, o.delivery_service, o.shardkey, o.sm_id, o.date_created, o.oof_shard,
        d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
        p.transaction, p.request_id, p.currency, p.provider, p.amount, p.payment_dt, p.bank, p.delivery_cost, p.goods_total, p.custom_fee,
        i.chrt_id, i.track_number AS item_track_number, i.price, i.rid, i.name AS item_name, i.sale, i.size, i.total_price, i.nm_id, i.brand, i.status
    FROM orders o
    LEFT JOIN deliveries d ON o.order_uid = d.order_uid
    LEFT JOIN payments p ON o.order_uid = p.order_uid
    LEFT JOIN items i ON o.order_uid = i.order_uid
    WHERE o.order_uid = $1`

	rows, err := s.pool.Query(ctx, sql, id)
	if err != nil {
		return models.Order{}, fmt.Errorf("failed to query order:%w", err)
	}
	defer rows.Close()

	var order models.Order
	// itemsMap := make(map[string][]models.Item)

	for rows.Next() {
		var (
			orderUID, trackNumber, entry, locale, internalSignature, customerID, deliveryService, shardkey, oofShard   string
			dName, dPhone, dZip, dCity, dAddress, dRegion, dEmail                                                      string
			pTransaction, pRequestID, pCurrency, pProvider, pBank                                                      string
			pPaymentDT                                                                                                 int64
			iTrackNumber, iRid, iName, iBrand, iSize                                                                   string
			iPrice, iSale, iTotalPrice, iNmID, iStatus, smID, pAmount, pDeliveryCost, pGoodsTotal, pCustomFee, iChrtID int
			dateCreated                                                                                                time.Time
		)

		err := rows.Scan(&orderUID, &trackNumber, &entry, &locale, &internalSignature, &customerID, &deliveryService, &shardkey, &smID, &dateCreated, &oofShard,
			&dName, &dPhone, &dZip, &dCity, &dAddress, &dRegion, &dEmail,
			&pTransaction, &pRequestID, &pCurrency, &pProvider, &pAmount, &pPaymentDT, &pBank, &pDeliveryCost, &pGoodsTotal, &pCustomFee,
			&iChrtID, &iTrackNumber, &iPrice, &iRid, &iName, &iSale, &iSize, &iTotalPrice, &iNmID, &iBrand, &iStatus)

		if err != nil {
			return models.Order{}, fmt.Errorf("failed to scan row:%w", &err)
		}
		if orderUID == "" {
			order = models.Order{
				OrderUID:          orderUID,
				TrackNumber:       trackNumber,
				Entry:             entry,
				Locale:            locale,
				InternalSignature: internalSignature,
				CustomerID:        customerID,
				DeliveryService:   deliveryService,
				ShardKey:          shardkey,
				SmID:              smID,
				DateCreated:       dateCreated,
				OofShard:          oofShard,
				Delivery: models.Delivery{
					Name:    dName,
					Phone:   dPhone,
					Zip:     dZip,
					City:    dCity,
					Address: dAddress,
					Region:  dRegion,
					Email:   dEmail,
				},
				Payment: models.Payment{
					Transaction:  pTransaction,
					RequestID:    pRequestID,
					Currency:     pCurrency,
					Provider:     pProvider,
					Amount:       pAmount,
					PaymentDT:    pPaymentDT,
					Bank:         pBank,
					DeliveryCost: pDeliveryCost,
					GoodsTotal:   pGoodsTotal,
					CustomFee:    pCustomFee,
				},
			}
		}
		item := models.Item{
			ChrtID:      iChrtID,
			TrackNumber: iTrackNumber,
			Price:       iPrice,
			Rid:         iRid,
			Name:        iName,
			Sale:        iSale,
			Size:        iSize,
			TotalPrice:  iTotalPrice,
			NmID:        iNmID,
			Brand:       iBrand,
			Status:      iStatus,
		}
		order.Items = append(order.Items, item)

	}

	if err = rows.Err(); err != nil {
		return models.Order{}, fmt.Errorf("rows error: %w", err)
	}

	return order, nil
}

// func (s *OrderStoragePostgres) GetLastOrders(ctx context.Context, limit int) ([]models.Order, error) {

// }

func saveOrder(ctx context.Context, order models.Order, q Queryer) error {
	sql := `INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	_, err := q.Exec(ctx, sql, order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature, order.CustomerID, order.ShardKey, order.SmID, order.DateCreated, order.OofShard)
	if err != nil {
		return err
	}
	return nil
}

func saveDelivery(ctx context.Context, q Queryer, delivery models.Delivery, orderID string) error {
	sql := `INSERT INTO deliveries(order_uid, name, phone, zip, city, address, region,email)`
	_, err := q.Exec(ctx, sql, orderID, delivery.Name, delivery.Phone, delivery.Zip, delivery.City, delivery.Address, delivery.Region, delivery.Email)
	if err != nil {
		return err
	}
	return nil
}
func savePayment(ctx context.Context, q Queryer, payment models.Payment, orderID string) error {
	sql := `INSERT INTO payments(order_uid,transaction,request_id,currency,provider,amount ,payment_dt,bank ,delivery_cost,goods_total ,custom_fee )`
	_, err := q.Exec(ctx, sql, orderID, payment.Transaction, payment.RequestID, payment.Currency, payment.Provider, payment.Amount, payment.PaymentDT, payment.Bank, payment.DeliveryCost, payment.GoodsTotal, payment.CustomFee)
	if err != nil {
		return err
	}
	return nil
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
