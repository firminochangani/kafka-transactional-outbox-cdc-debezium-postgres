-- +goose Up
CREATE SCHEMA orders_svc;

CREATE TABLE orders_svc.orders (
    id TEXT PRIMARY KEY,
    customer_id TEXT NOT NULL,
    placed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE orders_svc.order_items (
    id TEXT PRIMARY KEY,
    product_id TEXT,
    order_id TEXT NOT NULL,
    name TEXT NOT NULL,
    price BIGINT NOT NULL,
    quantity INT NOT NULL,

    CONSTRAINT un_product_id_order_id UNIQUE(product_id, order_id),
    CONSTRAINT fk_order_id FOREIGN KEY (order_id) REFERENCES orders_svc.orders(id)
);

CREATE TABLE orders_svc.outbox (
    id            BIGSERIAL PRIMARY KEY,
    aggregateid   TEXT NOT NULL,
    aggregatetype TEXT NOT NULL,
    type          TEXT NOT NULL,
    payload       JSONB NOT NULL
);

-- +goose Down
DROP TABLE orders_svc.order_items;
DROP TABLE orders_svc.orders;
DROP TABLE orders_svc.outbox;
DROP SCHEMA orders_svc;
