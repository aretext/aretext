-- 18. Analytical functions with frame clauses
SELECT 
  department_id,
  employee_id,
  full_name,
  hire_date,
  salary,
  FIRST_VALUE(full_name) OVER dept_window as highest_paid_in_dept,
  LAST_VALUE(full_name) OVER dept_window as lowest_paid_in_dept,
  LEAD(full_name, 1, 'None') OVER dept_hire_window as next_hire_name,
  LEAD(hire_date, 1) OVER dept_hire_window as next_hire_date,
  LAG(full_name, 1, 'None') OVER dept_hire_window as prev_hire_name,
  salary - LAG(salary, 1, 0) OVER dept_hire_window as salary_diff_from_prev,
  CUME_DIST() OVER dept_salary_window as salary_percentile,
  NTILE(4) OVER dept_salary_window as salary_quartile
FROM employees
WHERE status = 'Active'
WINDOW
  dept_window AS (PARTITION BY department_id ORDER BY salary DESC ROWS BETWEEN UNBOUNDED PRECEDING AND UNBOUNDED FOLLOWING),
  dept_hire_window AS (PARTITION BY department_id ORDER BY hire_date ASC),
  dept_salary_window AS (PARTITION BY department_id ORDER BY salary DESC)
ORDER BY department_id, salary DESC;
