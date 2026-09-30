WITH RECURSIVE edges AS (
  SELECT a, b FROM friendships
  UNION
  SELECT b, a FROM friendships
),
reach AS (
  SELECT id AS start, id AS node
  FROM users

  UNION                                   -- не UNION ALL: повтор пары обрывает цикл
  SELECT r.start, e.b
  FROM reach r
  JOIN edges e ON e.a = r.node
),
component AS (
  SELECT start AS user_id, min(node) AS component
  FROM reach
  GROUP BY start
)
SELECT c.component,
       count(*)                                  AS size,
       string_agg(u.name, ', ' ORDER BY u.name)  AS members
FROM component c
JOIN users u ON u.id = c.user_id
GROUP BY c.component
ORDER BY c.component;
