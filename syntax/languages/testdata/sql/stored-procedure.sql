-- 14. Stored procedure with transaction, error handling, and dynamic SQL
CREATE OR REPLACE PROCEDURE transfer_funds(
  sender_id INT,
  receiver_id INT,
  amount DECIMAL(10,2)
)
LANGUAGE plpgsql
AS $$
DECLARE
  sender_balance DECIMAL(10,2);
  transfer_fee DECIMAL(10,2);
  transaction_id INT;
  sql_stmt TEXT;
BEGIN
  -- Start transaction
  BEGIN
    -- Get sender's balance
    SELECT current_balance INTO sender_balance FROM accounts WHERE account_id = sender_id FOR UPDATE;
    
    IF sender_balance < amount THEN
      RAISE EXCEPTION 'Insufficient funds: balance is %', sender_balance;
    END IF;
    
    -- Calculate transfer fee
    transfer_fee := CASE 
      WHEN amount < 1000 THEN 5.00
      WHEN amount < 10000 THEN 15.00
      ELSE amount * 0.01
    END;
    
    -- Update sender account
    UPDATE accounts 
    SET current_balance = current_balance - (amount + transfer_fee),
        last_transaction_date = CURRENT_TIMESTAMP
    WHERE account_id = sender_id;
    
    -- Update receiver account
    UPDATE accounts 
    SET current_balance = current_balance + amount,
        last_transaction_date = CURRENT_TIMESTAMP
    WHERE account_id = receiver_id;
    
    -- Insert transaction record
    INSERT INTO transactions (sender_id, receiver_id, amount, fee, transaction_date, status)
    VALUES (sender_id, receiver_id, amount, transfer_fee, CURRENT_TIMESTAMP, 'Completed')
    RETURNING transaction_id INTO transaction_id;
    
    -- Create audit log with dynamic SQL
    sql_stmt := 'INSERT INTO audit_logs (entity_type, entity_id, action, details, user_id, timestamp) VALUES ($1, $2, $3, $4, $5, $6)';
    EXECUTE sql_stmt USING 'transaction', transaction_id, 'transfer', 
      format('{"amount": %s, "sender": %s, "receiver": %s}', amount, sender_id, receiver_id),
      current_user, CURRENT_TIMESTAMP;
    
    -- Commit transaction
    COMMIT;
  EXCEPTION WHEN OTHERS THEN
    -- Rollback on error
    ROLLBACK;
    RAISE;
  END;
END;
$$;
