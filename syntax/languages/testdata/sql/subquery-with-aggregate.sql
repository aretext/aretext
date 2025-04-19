-- 2. Subquery in WHERE clause with aggregate functions
SELECT department_id, department_name
FROM departments
WHERE department_id IN (
  SELECT department_id
  FROM employees
  GROUP BY department_id
  HAVING COUNT(*) > 10 AND AVG(salary) > 50000
);
