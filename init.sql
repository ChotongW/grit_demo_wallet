CREATE TABLE IF NOT EXISTS ledger_entries (
    id VARCHAR(36) PRIMARY KEY,
    transaction_id VARCHAR(36) NOT NULL,
    account_id VARCHAR(50) NOT NULL,
    amount NUMERIC(20, 2) NOT NULL,
    direction VARCHAR(10) NOT NULL CHECK (direction IN ('DEBIT', 'CREDIT')),
    reference_id VARCHAR(255),
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT unq_ledger_tx_account UNIQUE (transaction_id, account_id)
);

CREATE INDEX IF NOT EXISTS idx_ledger_account_id ON ledger_entries(account_id);
CREATE INDEX IF NOT EXISTS idx_ledger_transaction_id ON ledger_entries(transaction_id);
CREATE INDEX IF NOT EXISTS idx_ledger_created_at ON ledger_entries(created_at);
CREATE INDEX IF NOT EXISTS idx_ledger_reference_id ON ledger_entries(reference_id);

CREATE TABLE IF NOT EXISTS accounts (
    account_id VARCHAR(50) PRIMARY KEY,
    account_type VARCHAR(20) NOT NULL CHECK (account_type IN ('SYSTEM', 'USER')),
    user_id VARCHAR(36),
    email VARCHAR(255) UNIQUE,
    referrer_account_id VARCHAR(50),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE balances (
    account_id VARCHAR(50) PRIMARY KEY, 
    amount NUMERIC(20, 2) NOT NULL DEFAULT 0,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    CONSTRAINT fk_account 
      FOREIGN KEY (account_id) 
      REFERENCES accounts(account_id) 
      ON DELETE CASCADE
) WITH (fillfactor = 70);

CREATE INDEX IF NOT EXISTS idx_accounts_user_id ON accounts(user_id);
CREATE INDEX IF NOT EXISTS idx_accounts_email ON accounts(email);
CREATE INDEX IF NOT EXISTS idx_accounts_referrer ON accounts(referrer_account_id);

INSERT INTO accounts (account_id, account_type, created_at)
VALUES ('1001', 'SYSTEM', NOW())
ON CONFLICT (account_id) DO NOTHING;

INSERT INTO balances (account_id, amount, updated_at)
VALUES ('1001', 10000.00, NOW())
ON CONFLICT (account_id) DO NOTHING;

INSERT INTO accounts (account_id, account_type, created_at)
VALUES ('1002', 'SYSTEM', NOW())
ON CONFLICT (account_id) DO NOTHING;

INSERT INTO balances (account_id, amount, updated_at)
VALUES ('1002', 0.00, NOW())
ON CONFLICT (account_id) DO NOTHING;

INSERT INTO accounts (account_id, account_type, created_at)
VALUES ('1003', 'SYSTEM', NOW())
ON CONFLICT (account_id) DO NOTHING;

INSERT INTO balances (account_id, amount, updated_at)
VALUES ('1003', 0.00, NOW())
ON CONFLICT (account_id) DO NOTHING;

INSERT INTO accounts (account_id, account_type, created_at)
VALUES ('1004', 'SYSTEM', NOW())
ON CONFLICT (account_id) DO NOTHING;

INSERT INTO balances (account_id, amount, updated_at)
VALUES ('1004', 0.00, NOW())
ON CONFLICT (account_id) DO NOTHING;

-- Create initial ledger entry for Referral Funding Pool (for audit trail)
-- INSERT INTO ledger_entries (id, transaction_id, account_id, amount, direction, reference_id, description, created_at)
-- SELECT
--     gen_random_uuid()::text,
--     gen_random_uuid()::text,
--     '1001',
--     10000.00,
--     'CREDIT',
--     'INIT-REFERRAL-POOL',
--     'Initial funding for referral rewards pool',
--     NOW()
-- WHERE NOT EXISTS (
--     SELECT 1 FROM ledger_entries WHERE reference_id = 'INIT-REFERRAL-POOL'
-- );
