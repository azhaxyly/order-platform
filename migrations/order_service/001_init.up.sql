
CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL,
    total_amount BIGINT NOT NULL, 
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);


CREATE TABLE IF NOT EXISTS outbox (
    id UUID PRIMARY KEY,
    aggregate_id UUID NOT NULL,   
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,       
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    published_at TIMESTAMP WITH TIME ZONE, 
    retry_count INT DEFAULT 0
);


CREATE INDEX IF NOT EXISTS idx_outbox_unpublished 
ON outbox (created_at) 
WHERE published_at IS NULL;