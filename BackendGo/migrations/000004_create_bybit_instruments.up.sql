CREATE TABLE IF NOT EXISTS bybit_instruments (
    symbol VARCHAR(20) PRIMARY KEY,
    category VARCHAR(20) NOT NULL,
    base_coin VARCHAR(10) NOT NULL,
    quote_coin VARCHAR(10) NOT NULL,
    price_precision INTEGER NOT NULL,
    quantity_precision INTEGER NOT NULL,
    min_price DECIMAL(40,20) NOT NULL,
    max_price DECIMAL(40,20) NOT NULL,
    min_quantity DECIMAL(40,20) NOT NULL,
    max_quantity DECIMAL(40,20) NOT NULL,
    quantity_step DECIMAL(40,20) NOT NULL,
    price_step DECIMAL(40,20) NOT NULL,
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_bybit_instruments_symbol ON bybit_instruments(symbol);
CREATE INDEX idx_bybit_instruments_category ON bybit_instruments(category);
CREATE INDEX idx_bybit_instruments_base_coin ON bybit_instruments(base_coin);
CREATE INDEX idx_bybit_instruments_quote_coin ON bybit_instruments(quote_coin); 