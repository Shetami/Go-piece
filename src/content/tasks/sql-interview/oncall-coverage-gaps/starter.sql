-- Дыры в графике дежурств за неделю 2024-09-02 — 2024-09-09.
-- Колонки: team, gap_start, gap_end, gap_hours. Порядок: team, gap_start.
WITH ordered AS (
  SELECT team_id, starts_at,
         lag(ends_at) OVER (PARTITION BY team_id ORDER BY starts_at) AS prev_end
  FROM oncall_shifts
)
SELECT t.name AS team, o.prev_end AS gap_start, o.starts_at AS gap_end,
       round(extract(epoch FROM o.starts_at - o.prev_end) / 3600, 1) AS gap_hours
FROM ordered o
JOIN teams t ON t.id = o.team_id
WHERE o.starts_at > o.prev_end
ORDER BY team, gap_start;
