CREATE TABLE IF NOT EXISTS orders (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending',
    total_amount    NUMERIC(12, 2) NOT NULL,
    payment_method  TEXT NOT NULL,
    delivery_city   TEXT NOT NULL,
    delivery_street TEXT NOT NULL,
    delivery_zip    TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_status  ON orders(status);
