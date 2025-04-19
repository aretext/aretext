-- 17. Full-text search with ranking
SELECT 
  product_id,
  product_name,
  category,
  price,
  ts_rank_cd(to_tsvector('english', product_name || ' ' || description), query) as relevance
FROM 
  products,
  to_tsquery('english', 'portable & wireless & !headphones') as query
WHERE 
  to_tsvector('english', product_name || ' ' || description) @@ query
  AND is_active = TRUE
  AND inventory_count > 0
ORDER BY relevance DESC, price ASC
LIMIT 20;
