WITH params AS (
  SELECT timestamp '2024-09-02 00:00' AS week_start,
         timestamp '2024-09-09 00:00' AS week_end
),
clipped AS (
  SELECT s.team_id,
         greatest(s.starts_at, p.week_start) AS starts_at,
         least(s.ends_at, p.week_end)        AS ends_at
  FROM oncall_shifts s, params p
  WHERE NOT s.is_cancelled
    AND s.starts_at < p.week_end
    AND s.ends_at   > p.week_start
  UNION ALL
  -- «пустая смена» в конце недели: ловит хвост и команды без смен
  SELECT t.id, p.week_end, p.week_end
  FROM teams t, params p
),
covered AS (
  SELECT c.*,
         max(ends_at) OVER (PARTITION BY team_id ORDER BY starts_at, ends_at
                            ROWS BETWEEN UNBOUNDED PRECEDING AND 1 PRECEDING) AS covered_until
  FROM clipped c
)
SELECT t.name AS team,
       coalesce(c.covered_until, p.week_start) AS gap_start,
       c.starts_at AS gap_end,
       round(extract(epoch FROM c.starts_at - coalesce(c.covered_until, p.week_start)) / 3600, 1) AS gap_hours
FROM covered c
JOIN teams t ON t.id = c.team_id
CROSS JOIN params p
WHERE c.starts_at > coalesce(c.covered_until, p.week_start)
ORDER BY team, gap_start;
