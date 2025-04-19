-- 16. Temporal query with time ranges and intervals
SELECT 
  e.employee_id,
  e.full_name,
  p.project_name,
  pa.assignment_start,
  pa.assignment_end,
  EXTRACT(EPOCH FROM (pa.assignment_end - pa.assignment_start)) / 86400 as days_assigned,
  (CASE 
    WHEN pa.assignment_end > CURRENT_DATE THEN 'Active'
    ELSE 'Completed'
  END) as status,
  (pa.assignment_end - pa.assignment_start) * p.daily_rate as total_project_cost
FROM employees e
JOIN project_assignments pa ON e.employee_id = pa.employee_id
JOIN projects p ON pa.project_id = p.project_id
WHERE 
  (pa.assignment_start, pa.assignment_end) OVERLAPS (DATE '2023-01-01', DATE '2023-12-31')
  AND NOT EXISTS (
    SELECT 1 FROM employee_vacations v
    WHERE e.employee_id = v.employee_id
    AND (v.start_date, v.end_date) OVERLAPS (pa.assignment_start, pa.assignment_end)
  )
ORDER BY pa.assignment_start;
