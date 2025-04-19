-- 3. Common Table Expression (CTE) with recursive query
WITH RECURSIVE employee_hierarchy AS (
  SELECT employee_id, manager_id, full_name, 1 AS level
  FROM employees
  WHERE manager_id IS NULL
  
  UNION ALL
  
  SELECT e.employee_id, e.manager_id, e.full_name, eh.level + 1
  FROM employees e
  JOIN employee_hierarchy eh ON e.manager_id = eh.employee_id
)
SELECT * FROM employee_hierarchy
ORDER BY level, full_name;
