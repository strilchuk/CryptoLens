CREATE TABLE IF NOT EXISTS bybit_instruments (
    symbol VARCHAR(20) PRIMARY KEY,
    category VARCHAR(20) NOT NULL,
    base_coin VARCHAR(10) NOT NULL,
    quote_coin VARCHAR(10) NOT NULL,
    innovation VARCHAR(10) NOT NULL,
    status VARCHAR(20) NOT NULL,
    margin_trading VARCHAR(20) NOT NULL,
    st_tag VARCHAR(10) NOT NULL,
    base_precision DECIMAL(65,30) NOT NULL,
    quote_precision DECIMAL(65,30) NOT NULL,
    min_order_qty DECIMAL(65,30) NOT NULL,
    max_order_qty DECIMAL(65,30) NOT NULL,
    min_order_amt DECIMAL(65,30) NOT NULL,
    max_order_amt DECIMAL(65,30) NOT NULL,
    tick_size DECIMAL(65,30) NOT NULL,
    price_limit_ratio_x DECIMAL(65,30) NOT NULL,
    price_limit_ratio_y DECIMAL(65,30) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_bybit_instruments_symbol ON bybit_instruments(symbol);
CREATE INDEX idx_bybit_instruments_category ON bybit_instruments(category);
CREATE INDEX idx_bybit_instruments_base_coin ON bybit_instruments(base_coin);
CREATE INDEX idx_bybit_instruments_quote_coin ON bybit_instruments(quote_coin); 