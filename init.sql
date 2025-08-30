CREATE DATABASE wb_l0;

CREATE TABLE IF NOT EXISTS orders (
    order_uid VARCHAR(255) PRIMARY KEY,
    track_number VARCHAR(255),
    entry VARCHAR(255),
    locale VARCHAR(10),
    internal_signature VARCHAR(255),
    customer_id VARCHAR(255),
    delivery_service VARCHAR(255),
    shardkey VARCHAR(10),
    sm_id INT,
    date_created TIMESTAMP WITH TIME ZONE,
    oof_shard VARCHAR(10)
    );

CREATE TABLE IF NOT EXISTS delivery (
    order_uid VARCHAR(255) PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE,
    fio VARCHAR(255),
    phone VARCHAR(40),
    zip VARCHAR(40),
    city VARCHAR(255),
    address VARCHAR(255),
    region VARCHAR(255),
    email VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS payment (
    order_uid VARCHAR(255) PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE,
    transaction_number VARCHAR(255),
    request_id VARCHAR(255),
    currency VARCHAR(10),
    provider VARCHAR(50),
    amount INT,
    payment_dt BIGINT,
    bank VARCHAR(255),
    delivery_cost INT,
    goods_total INT,
    custom_fee INT
);

CREATE TABLE IF NOT EXISTS items (
    id SERIAL PRIMARY KEY,
    order_uid VARCHAR(255) REFERENCES orders(order_uid) ON DELETE CASCADE,
    chrt_id INT,
    track_number VARCHAR(255),
    price INT,
    rid VARCHAR(255),
    item_name VARCHAR(255),
    sale INT,
    item_size VARCHAR(10),
    total_price INT,
    nm_id INT,
    brand VARCHAR(255),
    status INT
);

CREATE USER user1 WITH LOGIN PASSWORD '123456789';

GRANT CONNECT ON DATABASE wb_l01 TO user1;

GRANT USAGE ON SCHEMA public TO user1;

GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO user1;

GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO user1;