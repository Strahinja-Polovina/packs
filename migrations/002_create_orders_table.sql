-- +goose Up
CREATE TABLE orders (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    package_size INTEGER NOT NULL CHECK (package_size > 0),
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(order_id, package_size)
);

CREATE INDEX idx_order_items_order_id ON order_items(order_id);

-- +goose Down
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;