-- 5. UNION, INTERSECT, and EXCEPT operations
SELECT customer_id, 'High Value' as customer_type
FROM customers
WHERE lifetime_value > 10000

UNION

SELECT customer_id, 'Recent' as customer_type
FROM orders
WHERE order_date > CURRENT_DATE - INTERVAL '30 days'

EXCEPT

SELECT customer_id, 'Recent' as customer_type
FROM customers
WHERE status = 'Inactive';
