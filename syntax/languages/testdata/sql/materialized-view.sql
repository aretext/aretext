-- 15. Materialized view with complex joins and window functions
CREATE MATERIALIZED VIEW monthly_sales_summary AS
SELECT 
  DATE_TRUNC('month', s.sale_date) as sale_month,
  p.category,
  r.region_name,
  COUNT(*) as transaction_count,
  COUNT(DISTINCT s.customer_id) as unique_customers,
  SUM(s.quantity) as total_units_sold,
  SUM(s.amount) as total_revenue,
  AVG(s.amount) as avg_transaction_value,
  PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY s.amount) as median_transaction_value,
  SUM(s.amount) - SUM(p.cost * s.quantity) as gross_profit,
  (SUM(s.amount) - SUM(p.cost * s.quantity)) / NULLIF(SUM(s.amount), 0) * 100 as profit_margin,
  LAG(SUM(s.amount)) OVER (PARTITION BY p.category, r.region_name ORDER BY DATE_TRUNC('month', s.sale_date)) as prev_month_revenue,
  SUM(s.amount) / NULLIF(LAG(SUM(s.amount)) OVER (PARTITION BY p.category, r.region_name ORDER BY DATE_TRUNC('month', s.sale_date)), 0) - 1 as month_over_month_growth
FROM sales s
JOIN products p ON s.product_id = p.product_id
JOIN regions r ON s.region_id = r.region_id
WHERE s.sale_date >= '2022-01-01'
GROUP BY DATE_TRUNC('month', s.sale_date), p.category, r.region_name
WITH DATA;

CREATE UNIQUE INDEX ON monthly_sales_summary (sale_month, category, region_name);
