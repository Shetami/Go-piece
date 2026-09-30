WITH r AS (
  SELECT a.name, count(*) AS resolved
  FROM tickets t
  JOIN agents a ON a.id = t.agent_id
  WHERE lower(t.status) = 'resolved'
    AND t.closed_at >= '2024-03-01' AND t.closed_at < '2024-04-01'   -- полуоткрытый интервал
  GROUP BY a.id, a.name
  ORDER BY resolved DESC            -- только по метрике: по ней и определяются «ничьи»
  FETCH FIRST 3 ROWS WITH TIES
)
SELECT name, resolved
FROM r
ORDER BY resolved DESC, name;       -- стабильный порядок вывода
