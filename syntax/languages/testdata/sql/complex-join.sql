-- 11. Complex JOIN with GROUP BY, HAVING, and ROLLUP
SELECT 
  COALESCE(c.category, 'All Categories') as category,
  COALESCE(r.region, 'All Regions') as region,
  EXTRACT(YEAR FROM o.order_date) as year,
  COUNT(DISTINCT o.customer_id) as unique_customers,
  SUM(oi.quantity * oi.unit_price) as total_revenue
FROM orders o
JOIN order_items oi ON o.order_id = oi.order_id
JOIN products p ON oi.product_id = p.product_id
JOIN product_categories c ON p.category_id = c.category_id
JOIN customer_regions r ON o.customer_id = r.customer_id
WHERE o.status = 'Completed'
  AND o.order_date BETWEEN '2022-01-01' AND '2023-12-31'
GROUP BY ROLLUP (c.category, r.region, EXTRACT(YEAR FROM o.order_date))
HAVING SUM(oi.quantity * oi.unit_price) > 10000
ORDER BY category NULLS FIRST, region NULLS FIRST, year NULLS FIRST;
