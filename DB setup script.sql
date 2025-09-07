CREATE TABLE orders (
    order_uid VARCHAR(50) PRIMARY KEY,
    track_number VARCHAR(50) NOT NULL,
    entry VARCHAR(10) NOT NULL,
    locale VARCHAR(5) NOT NULL,
    internal_signature VARCHAR(50),
    customer_id VARCHAR(50) NOT NULL,
    delivery_service VARCHAR(50) NOT NULL,
    shardkey VARCHAR(10) NOT NULL,
    sm_id INT NOT NULL,
    date_created TIMESTAMPTZ NOT NULL,
    oof_shard VARCHAR(10) NOT NULL
);

CREATE TABLE deliveries (
    id SERIAL PRIMARY KEY,
    order_uid VARCHAR(50) UNIQUE NOT NULL REFERENCES orders(order_uid) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    zip VARCHAR(20) NOT NULL,
    city VARCHAR(100) NOT NULL,
    address TEXT NOT NULL,
    region VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL
);

CREATE TABLE payments (
    id SERIAL PRIMARY KEY,
    order_uid VARCHAR(50) UNIQUE NOT NULL REFERENCES orders(order_uid) ON DELETE CASCADE,
    transaction_id VARCHAR(50) UNIQUE,
    request_id VARCHAR(50),
    currency VARCHAR(3) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    amount NUMERIC(10,2) NOT NULL,
    payment_dt TIMESTAMPTZ NOT NULL,
    bank VARCHAR(50) NOT NULL,
    delivery_cost NUMERIC(10,2) NOT NULL,
    goods_total NUMERIC(10,2) NOT NULL,
    custom_fee NUMERIC(10,2) NOT NULL
);

CREATE TABLE items (
    id SERIAL PRIMARY KEY,
    order_uid VARCHAR(50) NOT NULL REFERENCES orders(order_uid) ON DELETE CASCADE,
    chrt_id BIGINT NOT NULL,
    track_number VARCHAR(50) NOT NULL,
    price NUMERIC(10,2) NOT NULL,
    rid VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    sale INT NOT NULL,
    size VARCHAR(10) NOT NULL,
    total_price NUMERIC(10,2) NOT NULL,
    nm_id BIGINT NOT NULL,
    brand VARCHAR(100) NOT NULL,
    status INT NOT NULL
);
