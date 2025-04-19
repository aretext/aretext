-- 1. Basic SELECT with multiple joins and conditions
SELECT c.customer_name, o.order_date, p.product_name, oi.quantity
FROM customers c
INNER JOIN orders o ON c.customer_id = o.customer_id
LEFT JOIN order_items oi ON o.order_id = oi.order_id
RIGHT JOIN products p ON oi.product_id = p.product_id
WHERE o.order_date BETWEEN '2023-01-01' AND '2023-12-31'
  AND p.category IN ('Electronics', 'Books')
ORDER BY o.order_date DESC, c.customer_name ASC
LIMIT 100 OFFSET 20;
