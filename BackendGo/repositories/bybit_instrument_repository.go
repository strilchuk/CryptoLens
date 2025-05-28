package repositories

import (
	"CryptoLens_Backend/logger"
	"CryptoLens_Backend/models"
	"context"
	"database/sql"
)

type BybitInstrumentRepository struct {
	db *sql.DB
}

func NewBybitInstrumentRepository(db *sql.DB) *BybitInstrumentRepository {
	return &BybitInstrumentRepository{db: db}
}

func (r *BybitInstrumentRepository) SaveInstruments(ctx context.Context, instruments []models.BybitInstrument) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		logger.LogError("Failed to begin transaction: %v", err)
		return err
	}
	defer tx.Rollback()

	// Сначала удаляем все существующие инструменты
	_, err = tx.ExecContext(ctx, "DELETE FROM bybit_instruments")
	if err != nil {
		logger.LogError("Failed to delete existing instruments: %v", err)
		return err
	}

	// Подготавливаем запрос для вставки
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO bybit_instruments (
			symbol, category, base_coin, quote_coin,
			price_precision, quantity_precision,
			min_price, max_price, min_quantity, max_quantity,
			quantity_step, price_step, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`)
	if err != nil {
		logger.LogError("Failed to prepare insert statement: %v", err)
		return err
	}
	defer stmt.Close()

	// Вставляем новые инструменты
	for _, inst := range instruments {
		_, err = stmt.ExecContext(ctx,
			inst.Symbol,
			inst.Category,
			inst.BaseCoin,
			inst.QuoteCoin,
			inst.PricePrecision,
			inst.QuantityPrecision,
			inst.MinPrice,
			inst.MaxPrice,
			inst.MinQuantity,
			inst.MaxQuantity,
			inst.QuantityStep,
			inst.PriceStep,
			inst.Status,
		)
		if err != nil {
			logger.LogError("Failed to insert instrument %s: %v", inst.Symbol, err)
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		logger.LogError("Failed to commit transaction: %v", err)
		return err
	}

	logger.LogInfo("Successfully saved %d instruments to database", len(instruments))
	return nil
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
		logger.LogError("Failed to query instruments: %v", err)
		return nil, err
	}
	defer rows.Close()

	var instruments []models.BybitInstrument
	for rows.Next() {
		var inst models.BybitInstrument
		err := rows.Scan(
			&inst.Symbol,
			&inst.Category,
			&inst.BaseCoin,
			&inst.QuoteCoin,
			&inst.PricePrecision,
			&inst.QuantityPrecision,
			&inst.MinPrice,
			&inst.MaxPrice,
			&inst.MinQuantity,
			&inst.MaxQuantity,
			&inst.QuantityStep,
			&inst.PriceStep,
			&inst.Status,
		)
		if err != nil {
			logger.LogError("Failed to scan instrument row: %v", err)
			return nil, err
		}
		instruments = append(instruments, inst)
	}

	if err = rows.Err(); err != nil {
		logger.LogError("Error iterating instrument rows: %v", err)
		return nil, err
	}

	logger.LogInfo("Retrieved %d instruments from database", len(instruments))
	return instruments, nil
}
