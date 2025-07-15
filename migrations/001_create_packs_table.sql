-- +goose Up
CREATE TABLE packs (
    id UUID PRIMARY KEY,
    size INTEGER NOT NULL CHECK (size > 0) UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Insert some default pack sizes
INSERT INTO packs (id, size) VALUES 
    (gen_random_uuid(), 250),
    (gen_random_uuid(), 500),
    (gen_random_uuid(), 1000),
    (gen_random_uuid(), 2000),
    (gen_random_uuid(), 5000);

-- +goose Down
DROP TABLE IF EXISTS packs;
