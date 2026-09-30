WITH RECURSIVE explode AS (
  SELECT b.child_id AS part_id, b.qty * 5 AS qty
  FROM bom b
  WHERE b.parent_id = 1

  UNION ALL

  SELECT b.child_id, e.qty * b.qty
  FROM explode e
  JOIN bom b ON b.parent_id = e.part_id
)
SELECT p.name,
       p.unit,
       round(sum(e.qty), 3)               AS qty,
       round(sum(e.qty) * p.unit_cost, 2) AS cost
FROM explode e
JOIN parts p ON p.id = e.part_id
WHERE NOT EXISTS (SELECT 1 FROM bom b WHERE b.parent_id = e.part_id)
GROUP BY p.id, p.name, p.unit, p.unit_cost
ORDER BY cost DESC, p.name;
