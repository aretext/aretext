-- 19. Complex hierarchical query with multiple levels of nesting
WITH product_hierarchy AS (
  SELECT 
    p.product_id,
    p.product_name,
    p.price,
    pc.category_name,
    pc.parent_category_id,
    ARRAY[pc.category_id] as category_path,
    1 as level
  FROM products p
  JOIN product_categories pc ON p.category_id = pc.category_id
  WHERE pc.parent_category_id IS NULL
  
  UNION ALL
  
  SELECT 
    p.product_id,
    p.product_name,
    p.price,
    pc.category_name,
    pc.parent_category_id,
    ph.category_path || pc.category_id,
    ph.level + 1
  FROM products p
  JOIN product_categories pc ON p.category_id = pc.category_id
  JOIN product_hierarchy ph ON pc.parent_category_id = ph.category_path[array_length(ph.category_path, 1)]
  WHERE ph.level < 5
)
SELECT 
  ph.product_id,
  ph.product_name,
  ph.price,
  string_agg(c.category_name, ' > ' ORDER BY idx) as category_breadcrumbs,
  ph.level as category_depth
FROM product_hierarchy ph
CROSS JOIN LATERAL unnest(ph.category_path) WITH ORDINALITY AS u(category_id, idx)
JOIN product_categories c ON u.category_id = c.category_id
GROUP BY ph.product_id, ph.product_name, ph.price, ph.level
ORDER BY category_breadcrumbs, ph.product_name;
