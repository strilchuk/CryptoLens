CREATE TABLE IF NOT EXISTS trade_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    symbol VARCHAR(20) NOT NULL REFERENCES bybit_instruments(symbol) ON DELETE RESTRICT,
    exec_id VARCHAR(50) NOT NULL,
    order_id VARCHAR(50) NOT NULL,
    order_link_id VARCHAR(50),
    side VARCHAR(10) NOT NULL,
    exec_price NUMERIC(65,30) NOT NULL,
    exec_qty NUMERIC(65,30) NOT NULL,
    exec_fee NUMERIC(65,30) NOT NULL,
    fee_rate NUMERIC(10,8) NOT NULL,
    is_maker BOOLEAN NOT NULL,
    order_type VARCHAR(20) NOT NULL,
    exec_time TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (exec_id, user_id)
);

CREATE INDEX idx_trade_logs_user_id ON trade_logs (user_id);
CREATE INDEX idx_trade_logs_symbol ON trade_logs (symbol);
CREATE INDEX idx_trade_logs_exec_time ON trade_logs (exec_time); 