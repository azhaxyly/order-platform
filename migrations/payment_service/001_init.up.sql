CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL UNIQUE, -- ВАЖНО: UNIQUE гарантирует идемпотентность
    user_id UUID NOT NULL,
    amount BIGINT NOT NULL, -- Копейки/центы
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_payments_order_id ON payments(order_id);