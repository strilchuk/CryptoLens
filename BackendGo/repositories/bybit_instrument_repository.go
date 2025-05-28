package repositories

import (
	"context"
	"database/sql"
	"time"

	"CryptoLens_Backend/models"
)

type BybitInstrumentRepository struct {
	db *sql.DB
}

func NewBybitInstrumentRepository(db *sql.DB) *BybitInstrumentRepository {
	return &BybitInstrumentRepository{db: db}
}

func (r *BybitInstrumentRepository) UpsertInstruments(ctx context.Context, instruments []models.BybitInstrument) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Очищаем старые данные
	_, err = tx.ExecContext(ctx, "DELETE FROM bybit_instruments")
	if err != nil {
		return err
	}

	// Вставляем новые данные
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO bybit_instruments (
			symbol, category, base_coin, quote_coin, 
			min_order_qty, max_order_qty, min_price, max_price,
			price_scale, quantity_scale, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now()
	for _, instrument := range instruments {
		_, err = stmt.ExecContext(ctx,
			instrument.Symbol,
			instrument.Category,
			instrument.BaseCoin,
			instrument.QuoteCoin,
			instrument.MinOrderQty,
			instrument.MaxOrderQty,
			instrument.MinPrice,
			instrument.MaxPrice,
			instrument.PriceScale,
			instrument.QuantityScale,
			instrument.Status,
			now,
			now,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *BybitInstrumentRepository) GetInstruments(ctx context.Context, category string) ([]models.BybitInstrument, error) {
	query := "SELECT * FROM bybit_instruments"
	args := []interface{}{}

	if category != "" {
		query += " WHERE category = $1"
		args = append(args, category)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var instruments []models.BybitInstrument
	for rows.Next() {
		var instrument models.BybitInstrument
		err := rows.Scan(
			&instrument.ID,
			&instrument.Symbol,
			&instrument.Category,
			&instrument.BaseCoin,
			&instrument.QuoteCoin,
			&instrument.MinOrderQty,
			&instrument.MaxOrderQty,
			&instrument.MinPrice,
			&instrument.MaxPrice,
			&instrument.PriceScale,
			&instrument.QuantityScale,
			&instrument.Status,
			&instrument.CreatedAt,
			&instrument.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		instruments = append(instruments, instrument)
	}

	return instruments, nil
}
