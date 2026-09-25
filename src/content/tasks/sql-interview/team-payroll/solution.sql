WITH RECURSIVE active AS (
  SELECT id, name, manager_id
  FROM employees
  WHERE hired_at <= date '2024-06-01'
    AND (fired_at IS NULL OR fired_at >= date '2024-06-01')
),
current_salary AS (
  SELECT DISTINCT ON (employee_id) employee_id, amount
  FROM salaries
  WHERE valid_from <= date '2024-06-01'
  ORDER BY employee_id, valid_from DESC
),
-- Все пары «начальник — подчинённый на любом уровне» и расстояние между ними
chain AS (
  SELECT manager_id AS boss_id, id AS emp_id, 1 AS depth
  FROM active
  WHERE manager_id IS NOT NULL
  UNION ALL
  SELECT c.boss_id, a.id, c.depth + 1
  FROM chain c
  JOIN active a ON a.manager_id = c.emp_id
)
SELECT b.name AS manager,
       count(*) FILTER (WHERE c.depth = 1) AS direct_reports,
       count(*)                            AS team_size,
       sum(s.amount)                       AS team_payroll,
       max(c.depth)                        AS depth
FROM chain c
JOIN active b ON b.id = c.boss_id
LEFT JOIN current_salary s ON s.employee_id = c.emp_id
GROUP BY b.id, b.name
ORDER BY team_size DESC, manager;
