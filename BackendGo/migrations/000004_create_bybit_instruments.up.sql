CREATE TABLE IF NOT EXISTS bybit_instruments (
    id SERIAL PRIMARY KEY,
    symbol VARCHAR(50) NOT NULL,
    category VARCHAR(20) NOT NULL,
    base_coin VARCHAR(20) NOT NULL,
    quote_coin VARCHAR(20) NOT NULL,
    min_order_qty DECIMAL(20,8) NOT NULL,
    max_order_qty DECIMAL(20,8) NOT NULL,
    min_price DECIMAL(20,8) NOT NULL,
    max_price DECIMAL(20,8) NOT NULL,
    price_scale INTEGER NOT NULL,
    quantity_scale INTEGER NOT NULL,
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_bybit_instruments_symbol ON bybit_instruments(symbol);
CREATE INDEX idx_bybit_instruments_category ON bybit_instruments(category);
CREATE INDEX idx_bybit_instruments_base_coin ON bybit_instruments(base_coin);
CREATE INDEX idx_bybit_instruments_quote_coin ON bybit_instruments(quote_coin); 