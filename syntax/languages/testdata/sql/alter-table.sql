-- 10. ALTER TABLE with multiple operations
ALTER TABLE customers
  ADD COLUMN loyalty_points INTEGER DEFAULT 0,
  ADD COLUMN membership_level VARCHAR(20),
  ADD CONSTRAINT check_valid_email CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'),
  ALTER COLUMN last_purchase_date SET DEFAULT CURRENT_DATE,
  DROP COLUMN IF EXISTS legacy_id,
  RENAME COLUMN customer_phone TO contact_number;
