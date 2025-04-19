-- 12. Pivot-like query using CASE and aggregate functions
SELECT 
  p.product_name,
  SUM(CASE WHEN r.region = 'North America' THEN s.amount ELSE 0 END) as north_america_sales,
  SUM(CASE WHEN r.region = 'Europe' THEN s.amount ELSE 0 END) as europe_sales,
  SUM(CASE WHEN r.region = 'Asia' THEN s.amount ELSE 0 END) as asia_sales,
  SUM(CASE WHEN r.region = 'South America' THEN s.amount ELSE 0 END) as south_america_sales,
  SUM(CASE WHEN r.region = 'Africa' THEN s.amount ELSE 0 END) as africa_sales,
  SUM(CASE WHEN r.region = 'Oceania' THEN s.amount ELSE 0 END) as oceania_sales,
  SUM(s.amount) as total_sales,
  COUNT(DISTINCT s.customer_id) as unique_customers
FROM sales s
JOIN products p ON s.product_id = p.product_id
JOIN regions r ON s.region_id = r.region_id
WHERE s.sale_date BETWEEN '2023-01-01' AND '2023-12-31'
GROUP BY p.product_name
HAVING SUM(s.amount) > 50000
ORDER BY total_sales DESC;
