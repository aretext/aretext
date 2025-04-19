-- 4. Window functions and complex aggregation
SELECT 
  department_id,
  employee_id,
  salary,
  RANK() OVER (PARTITION BY department_id ORDER BY salary DESC) as salary_rank,
  SUM(salary) OVER (PARTITION BY department_id) as dept_total_salary,
  salary / SUM(salary) OVER (PARTITION BY department_id) * 100 as percentage_of_dept,
  AVG(salary) OVER (ORDER BY hire_date ROWS BETWEEN 2 PRECEDING AND 2 FOLLOWING) as moving_avg_salary
FROM employees
WHERE status = 'Active';
