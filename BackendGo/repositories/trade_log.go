package repositories

import (
	"CryptoLens_Backend/integration/bybit"
	"CryptoLens_Backend/types"
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// TradeLogRepository реализует интерфейс TradeLogRepositoryInterface
type TradeLogRepository struct {
	db *sql.DB
}

// NewTradeLogRepository создает новый репозиторий для работы с логами торговли
func NewTradeLogRepository(db *sql.DB) types.TradeLogRepositoryInterface {
	return &TradeLogRepository{db: db}
}

// SaveExecution сохраняет информацию об исполнении ордера
func (r *TradeLogRepository) SaveExecution(ctx context.Context, userID string, exec bybit.ExecutionMessage) error {
	execPrice, err := decimal.NewFromString(exec.ExecPrice)
	if err != nil {
		return fmt.Errorf("invalid exec_price: %w", err)
	}
	execQty, err := decimal.NewFromString(exec.ExecQty)
	if err != nil {
		return fmt.Errorf("invalid exec_qty: %w", err)
	}
	execFee, err := decimal.NewFromString(exec.ExecFee)
	if err != nil {
		return fmt.Errorf("invalid exec_fee: %w", err)
	}
	feeRate, err := decimal.NewFromString(exec.FeeRate)
	if err != nil {
		return fmt.Errorf("invalid fee_rate: %w", err)
	}
	execTime, err := time.Parse(time.RFC3339, exec.ExecTime)
	if err != nil {
		return fmt.Errorf("invalid exec_time: %w", err)
	}

	_, err = r.db.ExecContext(ctx,
		`INSERT INTO trade_logs (
			user_id, symbol, exec_id, order_id, order_link_id, side, 
			exec_price, exec_qty, exec_fee, fee_rate, is_maker, 
			order_type, exec_time
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		userID, exec.Symbol, exec.ExecID, exec.OrderID, exec.OrderLinkID, exec.Side,
		execPrice, execQty, execFee, feeRate, exec.IsMaker,
		exec.OrderType, execTime,
	)
	if err != nil {
		return fmt.Errorf("failed to save execution: %w", err)
	}
	return nil
}
