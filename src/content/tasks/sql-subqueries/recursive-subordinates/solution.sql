WITH RECURSIVE subordinates AS (
  SELECT id, name, 1 AS depth
  FROM employees
  WHERE manager_id = 2

  UNION ALL

  SELECT e.id, e.name, s.depth + 1
  FROM employees e
  JOIN subordinates s ON e.manager_id = s.id
)
SELECT name, depth
FROM subordinates
ORDER BY depth, name;
