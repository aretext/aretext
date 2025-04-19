-- 6. Complex INSERT with SELECT and CASE statements
INSERT INTO sales_summary (region, product_category, year, total_sales, performance_level)
SELECT 
  r.region_name,
  p.category,
  EXTRACT(YEAR FROM s.sale_date),
  SUM(s.amount),
  CASE 
    WHEN SUM(s.amount) > 1000000 THEN 'Excellent'
    WHEN SUM(s.amount) > 500000 THEN 'Good'
    WHEN SUM(s.amount) > 100000 THEN 'Average'
    ELSE 'Poor'
  END
FROM sales s
JOIN regions r ON s.region_id = r.region_id
JOIN products p ON s.product_id = p.product_id
WHERE s.sale_date BETWEEN '2023-01-01' AND '2023-12-31'
GROUP BY r.region_name, p.category, EXTRACT(YEAR FROM s.sale_date);
