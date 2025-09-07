-- Orders (root)
CREATE TABLE IF NOT EXISTS orders (
    order_uid           TEXT        PRIMARY KEY,
    track_number        TEXT        NOT NULL,
    entry               TEXT        NOT NULL,
    locale              TEXT        NOT NULL,
    internal_signature  TEXT        NOT NULL DEFAULT '',
    customer_id         TEXT        NOT NULL,
    delivery_service    TEXT        NOT NULL,
    shardkey            TEXT        NOT NULL,
    sm_id               INTEGER     NOT NULL,
    date_created        TIMESTAMPTZ NOT NULL,
    oof_shard           TEXT        NOT NULL,

    -- Useful integrity checks
    CHECK (sm_id >= 0)
);

-- One-to-one: Delivery per order
CREATE TABLE IF NOT EXISTS deliveries (
    order_uid   TEXT    PRIMARY KEY
        REFERENCES orders(order_uid) ON DELETE CASCADE,
    name        TEXT    NOT NULL,
    phone       TEXT    NOT NULL,
    zip         TEXT    NOT NULL,
    city        TEXT    NOT NULL,
    address     TEXT    NOT NULL,
    region      TEXT    NOT NULL,
    email       TEXT    NOT NULL
);

-- One-to-one: Payment per order
CREATE TABLE IF NOT EXISTS payments (
    order_uid       TEXT PRIMARY KEY
        REFERENCES orders(order_uid) ON DELETE CASCADE,

    transaction     TEXT    NOT NULL UNIQUE,
    request_id      TEXT    NOT NULL DEFAULT '',
    currency        TEXT    NOT NULL,
    provider        TEXT    NOT NULL,
    amount          INTEGER NOT NULL,
    payment_dt      BIGINT  NOT NULL,  -- epoch seconds (matches JSON)
    bank            TEXT    NOT NULL,
    delivery_cost   INTEGER NOT NULL,
    goods_total     INTEGER NOT NULL,
    custom_fee      INTEGER NOT NULL,

    CHECK (amount >= 0),
    CHECK (delivery_cost >= 0),
    CHECK (goods_total >= 0),
    CHECK (custom_fee >= 0)
);

-- One-to-many: Items per order
CREATE TABLE IF NOT EXISTS items (
    id              BIGSERIAL PRIMARY KEY,  -- surrogate key for convenience
    order_uid       TEXT    NOT NULL
        REFERENCES orders(order_uid) ON DELETE CASCADE,

    chrt_id         INTEGER NOT NULL,
    track_number    TEXT    NOT NULL,
    price           INTEGER NOT NULL,
    rid             TEXT    NOT NULL,
    name            TEXT    NOT NULL,
    sale            INTEGER NOT NULL,
    size            TEXT    NOT NULL,
    total_price     INTEGER NOT NULL,
    nm_id           INTEGER NOT NULL,
    brand           TEXT    NOT NULL,
    status          INTEGER NOT NULL,

    CHECK (price >= 0),
    CHECK (sale >= 0),
    CHECK (total_price >= 0),
    CHECK (nm_id >= 0),
    CHECK (status >= 0)
);

-- Indexes to speed up common lookups (with IF NOT EXISTS)
CREATE INDEX IF NOT EXISTS idx_orders_track_number   ON orders   (track_number);
CREATE INDEX IF NOT EXISTS idx_orders_customer_id    ON orders   (customer_id);
CREATE INDEX IF NOT EXISTS idx_orders_date_created   ON orders   (date_created);

CREATE INDEX IF NOT EXISTS idx_items_order_uid       ON items    (order_uid);
CREATE INDEX IF NOT EXISTS idx_items_track_number    ON items    (track_number);

-- Optional: if (order_uid, chrt_id) should be unique per order, uncomment:
-- CREATE UNIQUE INDEX IF NOT EXISTS uq_items_order_chrt ON items(order_uid, chrt_id);
