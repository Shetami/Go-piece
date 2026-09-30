WITH RECURSIVE chain AS (
  -- шаг 1 — непосредственный руководитель автора расхода
  SELECT x.id AS expense_id, e.manager_id AS manager_id, 1 AS step
  FROM expenses x
  JOIN employees e ON e.id = x.employee_id
  WHERE e.manager_id IS NOT NULL

  UNION ALL

  SELECT c.expense_id, e.manager_id, c.step + 1
  FROM chain c
  JOIN employees e ON e.id = c.manager_id
  WHERE e.manager_id IS NOT NULL
)
SELECT x.id,
       author.name AS employee,
       a.name      AS approver,
       a.step
FROM expenses x
JOIN employees author ON author.id = x.employee_id
LEFT JOIN LATERAL (
  SELECT m.name, c.step
  FROM chain c
  JOIN employees m ON m.id = c.manager_id
  WHERE c.expense_id = x.id
    AND m.approval_limit >= x.amount
  ORDER BY c.step
  LIMIT 1
) a ON true
ORDER BY x.id;
