package db

import (
	"L0WB/internal/models"
	"context"
	"database/sql"
	"fmt"
	"log"
)

// создает новый заказ в базе данных
func (w *WbDB) CreateOrder(ctx context.Context, order *models.Order) error {

	tx, err := w.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Printf("Failed to rollback transaction: %v", rollbackErr)
			}
		}
	}()

	stmtOrder, err := tx.PrepareContext(ctx, `
        INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, customer_id,
        delivery_service, shardkey, sm_id, date_created, oof_shard)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
    `)
	if err != nil {
		return fmt.Errorf("failed to prepare order statement: %w", err)
	}
	defer stmtOrder.Close()

	stmtDelivery, err := tx.PrepareContext(ctx, `
        INSERT INTO delivery (order_uid, fio, phone, zip, city, address, region, email)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `)
	if err != nil {
		return fmt.Errorf("failed to prepare delivery statement: %w", err)
	}
	defer stmtDelivery.Close()

	stmtPayment, err := tx.PrepareContext(ctx, `
        INSERT INTO payment (order_uid, transaction_number, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
    `)
	if err != nil {
		return fmt.Errorf("failed to prepare payment statement: %w", err)
	}
	defer stmtPayment.Close()

	stmtItem, err := tx.PrepareContext(ctx, `
        INSERT INTO items (order_uid, chrt_id, track_number, price, rid, item_name, sale, item_size, total_price, nm_id, brand, status)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
    `)
	if err != nil {
		return fmt.Errorf("failed to prepare item statement: %w", err)
	}
	defer stmtItem.Close()

	_, err = stmtOrder.ExecContext(ctx,
		order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature,
		order.CustomerID, order.DeliveryService, order.Shardkey, order.SmID, order.DateCreated, order.OofShard,
	)
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	_, err = stmtDelivery.ExecContext(ctx,
		order.OrderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip,
		order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email,
	)
	if err != nil {
		return fmt.Errorf("failed to insert delivery: %w", err)
	}

	_, err = stmtPayment.ExecContext(ctx,
		order.OrderUID, order.Payment.TransactionNumber, order.Payment.RequestID, order.Payment.Currency,
		order.Payment.Provider, order.Payment.Amount, order.Payment.PaymentDT, order.Payment.Bank,
		order.Payment.DeliveryCost, order.Payment.GoodsTotal, order.Payment.CustomFee,
	)
	if err != nil {
		return fmt.Errorf("failed to insert payment: %w", err)
	}

	for _, item := range order.Items {
		_, err = stmtItem.ExecContext(ctx,
			order.OrderUID, item.ChrtID, item.TrackNumber, item.Price, item.RID, item.ItemName,
			item.Sale, item.ItemSize, item.TotalPrice, item.NmID, item.Brand, item.Status,
		)
		if err != nil {
			return fmt.Errorf("failed to insert item: %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Order %s created successfully", order.OrderUID)
	return nil
}

// получает заказ из базы данных по orderUID
func (w *WbDB) GetOrder(ctx context.Context, orderUID string) (*models.Order, error) {
	sqlStatement := `
		SELECT order_uid, track_number, entry, locale, internal_signature, customer_id,
		delivery_service, shardkey, sm_id, date_created, oof_shard
		FROM orders
		WHERE order_uid = $1
	`
	row := w.QueryRowContext(ctx, sqlStatement, orderUID)
	order := &models.Order{}
	err := row.Scan(
		&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature,
		&order.CustomerID, &order.DeliveryService, &order.Shardkey, &order.SmID, &order.DateCreated, &order.OofShard,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order not found: %s", orderUID)
		}
		return nil, fmt.Errorf("failed to get order from database: %w", err)
	}

	sqlStatement = `
		SELECT fio, phone, zip, city, address, region, email
		FROM delivery
		WHERE order_uid = $1
	`
	row = w.QueryRowContext(ctx, sqlStatement, orderUID)
	err = row.Scan(
		&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip,
		&order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region, &order.Delivery.Email,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Delivery info not found for order %s", orderUID)
		} else {
			return nil, fmt.Errorf("failed to get delivery info: %w", err)
		}
	}

	sqlStatement = `
		SELECT transaction_number, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
		FROM payment
		WHERE order_uid = $1
	`

	row = w.QueryRowContext(ctx, sqlStatement, orderUID)
	err = row.Scan(
		&order.Payment.TransactionNumber, &order.Payment.RequestID, &order.Payment.Currency,
		&order.Payment.Provider, &order.Payment.Amount, &order.Payment.PaymentDT, &order.Payment.Bank,
		&order.Payment.DeliveryCost, &order.Payment.GoodsTotal, &order.Payment.CustomFee,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Payment info not found for order %s", orderUID)
		} else {
			return nil, fmt.Errorf("failed to get payment info: %w", err)
		}
	}

	sqlStatement = `
		SELECT chrt_id, track_number, price, rid, item_name, sale, item_size, total_price, nm_id, brand, status
		FROM items
		WHERE order_uid = $1
	`
	rows, err := w.QueryContext(ctx, sqlStatement, orderUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get items: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		item := models.Item{}
		err = rows.Scan(
			&item.ChrtID, &item.TrackNumber, &item.Price, &item.RID, &item.ItemName,
			&item.Sale, &item.ItemSize, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}
		order.Items = append(order.Items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate items: %w", err)
	}
	return order, nil
}
