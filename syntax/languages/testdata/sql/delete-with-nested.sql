-- 8. DELETE with nested subqueries
DELETE FROM order_items
WHERE order_id IN (
  SELECT order_id
  FROM orders
  WHERE customer_id IN (
    SELECT customer_id
    FROM customers
    WHERE status = 'Fraudulent' OR (
      created_date < CURRENT_DATE - INTERVAL '5 years' AND
      last_active_date < CURRENT_DATE - INTERVAL '2 years'
    )
  )
);
