-- 20. SELECT with LATERAL joins, conditional aggregates, and string functions
SELECT 
  c.customer_id,
  c.customer_name,
  REGEXP_REPLACE(c.phone_number, '(\d{3})(\d{3})(\d{4})', '(\1) \2-\3') as formatted_phone,
  COALESCE(c.email, 'No email provided') as contact_email,
  orders.total_orders,
  orders.last_order_date,
  CASE 
    WHEN orders.last_order_date > CURRENT_DATE - INTERVAL '90 days' THEN 'Active'
    WHEN orders.last_order_date > CURRENT_DATE - INTERVAL '180 days' THEN 'Recent'
    WHEN orders.last_order_date > CURRENT_DATE - INTERVAL '365 days' THEN 'Inactive'
    ELSE 'Dormant'
  END as customer_status,
  top_products.product_names,
  c.address->>'street' || ', ' || 
  c.address->>'city' || ', ' || 
  c.address->>'state' || ' ' || 
  c.address->>'zip' as full_address
FROM customers c
LEFT JOIN LATERAL (
  SELECT 
    COUNT(*) as total_orders,
    MAX(order_date) as last_order_date,
    SUM(CASE WHEN status = 'Completed' THEN total_amount ELSE 0 END) as total_spent,
    AVG(CASE WHEN status = 'Completed' THEN total_amount ELSE NULL END) as avg_order_value
  FROM orders o
  WHERE o.customer_id = c.customer_id
) orders ON true
LEFT JOIN LATERAL (
  SELECT 
    string_agg(p.product_name, ', ' ORDER BY COUNT(*) DESC) as product_names
  FROM orders o
  JOIN order_items oi ON o.order_id = oi.order_id
  JOIN products p ON oi.product_id = p.product_id
  WHERE o.customer_id = c.customer_id
  GROUP BY o.customer_id
  LIMIT 3
) top_products ON true
WHERE (c.status = 'Active' OR orders.last_order_date > CURRENT_DATE - INTERVAL '180 days')
  AND c.customer_name ~* '^[A-M]'
ORDER BY orders.total_spent DESC NULLS LAST
LIMIT 50;
