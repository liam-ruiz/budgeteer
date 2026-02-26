CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- a single plaid connection
CREATE TABLE plaid_items (
    plaid_item_id TEXT PRIMARY KEY,
    app_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plaid_access_token TEXT NOT NULL, -- encrypted or handled by Secrets Manager
    institution_name TEXT NOT NULL,
    plaid_cursor TEXT,                -- the bookmark for /transactions/sync
    last_synced_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- the actual
CREATE TABLE bank_accounts (
    plaid_account_id TEXT PRIMARY KEY,
    plaid_item_id TEXT NOT NULL REFERENCES plaid_items(plaid_item_id) ON DELETE CASCADE,
    account_name TEXT NOT NULL,
    official_name TEXT,
    account_type TEXT NOT NULL,               -- e.g., 'depository', 'credit'
    account_subtype TEXT,                     -- e.g., 'checking', 'savings'
    current_balance NUMERIC(19, 4) NOT NULL DEFAULT 0,
    available_balance NUMERIC(19, 4) NOT NULL DEFAULT 0,
    iso_currency_code TEXT NOT NULL DEFAULT 'USD',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE transactions (
    plaid_transaction_id TEXT PRIMARY KEY,
    plaid_account_id TEXT NOT NULL REFERENCES bank_accounts (plaid_account_id) ON DELETE CASCADE,
    transaction_date DATE NOT NULL,
    transaction_name TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT '',
    amount NUMERIC(12, 2) NOT NULL,
    pending BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE budgets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    app_user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    category TEXT NOT NULL,
    limit_amount NUMERIC(12, 2) NOT NULL,
    budget_period TEXT NOT NULL DEFAULT 'monthly',
    start_date DATE NOT NULL,
    end_date DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);