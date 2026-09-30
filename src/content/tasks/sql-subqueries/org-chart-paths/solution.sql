WITH RECURSIVE tree AS (
  SELECT id, name, 1 AS level,
         name AS path,
         ARRAY[name] AS sort_key,
         ARRAY[id]   AS ids
  FROM employees
  WHERE manager_id IS NULL

  UNION ALL

  SELECT e.id, e.name, t.level + 1,
         t.path || ' / ' || e.name,
         t.sort_key || e.name,
         t.ids || e.id
  FROM employees e
  JOIN tree t ON e.manager_id = t.id
)
SELECT t.name,
       t.level,
       t.path,
       (SELECT count(*) FROM tree d WHERE t.id = ANY (d.ids)) - 1 AS team_size
FROM tree t
ORDER BY t.sort_key;
