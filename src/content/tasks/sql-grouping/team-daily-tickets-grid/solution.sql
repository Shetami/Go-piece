WITH grid AS (
  SELECT t.id AS team_id, t.name, d.day::date AS day
  FROM teams t
  CROSS JOIN generate_series(date '2024-07-01', date '2024-07-07', interval '1 day') AS d(day)
),
daily AS (
  SELECT g.team_id, g.name, g.day, count(k.id) AS cnt
  FROM grid g
  LEFT JOIN tickets k
         ON k.team_id = g.team_id
        AND k.created_at >= g.day
        AND k.created_at <  g.day + 1
  GROUP BY g.team_id, g.name, g.day
),
ranked AS (
  SELECT *, row_number() OVER (PARTITION BY team_id ORDER BY cnt DESC, day) AS rn
  FROM daily
)
SELECT name AS team,
       sum(cnt)                                     AS tickets,
       count(*) FILTER (WHERE cnt = 0)              AS zero_days,
       max(day) FILTER (WHERE rn = 1 AND cnt > 0)   AS busiest_day,
       max(cnt)                                     AS busiest_count
FROM ranked
GROUP BY team_id, name
ORDER BY name;
