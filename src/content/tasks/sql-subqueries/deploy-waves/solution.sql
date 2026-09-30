WITH RECURSIVE walk AS (
  SELECT name AS root, name AS node, 1 AS depth,
         ARRAY[name] AS path, false AS is_cycle
  FROM services

  UNION ALL

  SELECT w.root, d.depends_on, w.depth + 1,
         w.path || d.depends_on,
         d.depends_on = ANY (w.path)
  FROM walk w
  JOIN deps d ON d.service = w.node
  WHERE NOT w.is_cycle
)
SELECT root AS service,
       CASE WHEN bool_or(is_cycle) THEN NULL ELSE max(depth) END AS wave
FROM walk
GROUP BY root
ORDER BY wave NULLS LAST, service;
