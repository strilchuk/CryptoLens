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

	// Подготавливаем запрос для вставки/обновления
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO bybit_instruments (
			symbol, category, base_coin, quote_coin,
			innovation, status, margin_trading, st_tag,
			base_precision, quote_precision,
			min_order_qty, max_order_qty,
			min_order_amt, max_order_amt,
			tick_size, price_limit_ratio_x, price_limit_ratio_y
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		ON CONFLICT (symbol) DO UPDATE SET
			category = EXCLUDED.category,
			base_coin = EXCLUDED.base_coin,
			quote_coin = EXCLUDED.quote_coin,
			innovation = EXCLUDED.innovation,
			status = EXCLUDED.status,
			margin_trading = EXCLUDED.margin_trading,
			st_tag = EXCLUDED.st_tag,
			base_precision = EXCLUDED.base_precision,
			quote_precision = EXCLUDED.quote_precision,
			min_order_qty = EXCLUDED.min_order_qty,
			max_order_qty = EXCLUDED.max_order_qty,
			min_order_amt = EXCLUDED.min_order_amt,
			max_order_amt = EXCLUDED.max_order_amt,
			tick_size = EXCLUDED.tick_size,
			price_limit_ratio_x = EXCLUDED.price_limit_ratio_x,
			price_limit_ratio_y = EXCLUDED.price_limit_ratio_y,
			updated_at = CURRENT_TIMESTAMP
	`)
	if err != nil {
		logger.LogError("Failed to prepare insert statement: %v", err)
		return err
	}
	defer stmt.Close()

	// Вставляем или обновляем инструменты
	for _, inst := range instruments {
		_, err = stmt.ExecContext(ctx,
			inst.Symbol,
			inst.Category,
			inst.BaseCoin,
			inst.QuoteCoin,
			inst.Innovation,
			inst.Status,
			inst.MarginTrading,
			inst.StTag,
			inst.BasePrecision,
			inst.QuotePrecision,
			inst.MinOrderQty,
			inst.MaxOrderQty,
			inst.MinOrderAmt,
			inst.MaxOrderAmt,
			inst.TickSize,
			inst.PriceLimitRatioX,
			inst.PriceLimitRatioY,
		)
		if err != nil {
			logger.LogError("Failed to insert/update instrument %s: %v", inst.Symbol, err)
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
			&inst.Innovation,
			&inst.Status,
			&inst.MarginTrading,
			&inst.StTag,
			&inst.BasePrecision,
			&inst.QuotePrecision,
			&inst.MinOrderQty,
			&inst.MaxOrderQty,
			&inst.MinOrderAmt,
			&inst.MaxOrderAmt,
			&inst.TickSize,
			&inst.PriceLimitRatioX,
			&inst.PriceLimitRatioY,
			&inst.CreatedAt,
			&inst.UpdatedAt,
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

// Exists проверяет существование инструмента по символу
func (r *BybitInstrumentRepository) Exists(ctx context.Context, symbol string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM bybit_instruments
			WHERE symbol = $1
		)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, symbol).Scan(&exists)
	if err != nil {
		logger.LogError("Failed to check instrument existence: %v", err)
		return false, err
	}

	return exists, nil
}
