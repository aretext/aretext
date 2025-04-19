-- 7. UPDATE with joins and subqueries
UPDATE products p
SET 
  price = p.price * 1.1,
  last_updated = CURRENT_TIMESTAMP,
  status = CASE 
            WHEN inventory_count > 100 THEN 'In Stock' 
            WHEN inventory_count > 0 THEN 'Low Stock'
            ELSE 'Out of Stock'
          END
FROM inventory i
WHERE p.product_id = i.product_id
  AND p.category IN (
    SELECT category FROM product_categories WHERE seasonal = TRUE
  )
  AND NOT EXISTS (
    SELECT 1 FROM promotions pr WHERE pr.product_id = p.product_id AND pr.active = TRUE
  );
