CREATE TABLE IF NOT EXISTS user_instruments (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    symbol VARCHAR(20) NOT NULL REFERENCES bybit_instruments(symbol) ON DELETE RESTRICT,
    is_active BOOLEAN DEFAULT true NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(user_id, symbol)
);

CREATE INDEX idx_user_instruments_user_id ON user_instruments(user_id);
CREATE INDEX idx_user_instruments_symbol ON user_instruments(symbol);
CREATE INDEX idx_user_instruments_is_active ON user_instruments(is_active); 