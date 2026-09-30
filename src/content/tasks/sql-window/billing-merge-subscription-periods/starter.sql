-- Склеить пересекающиеся и смежные периоды доступа каждого клиента в непрерывные отрезки.
-- Колонки: customer, starts_on, ends_on, days (включительно). Порядок: customer, starts_on.
WITH flagged AS (
  SELECT customer, starts_on, ends_on,
         CASE WHEN starts_on <= lag(ends_on) OVER w THEN 0 ELSE 1 END AS is_start
  FROM periods
  WINDOW w AS (PARTITION BY customer ORDER BY starts_on)
),
grouped AS (
  SELECT *, sum(is_start) OVER (PARTITION BY customer ORDER BY starts_on) AS grp
  FROM flagged
)
SELECT customer, min(starts_on) AS starts_on, max(ends_on) AS ends_on,
       max(ends_on) - min(starts_on) + 1 AS days
FROM grouped
GROUP BY customer, grp
ORDER BY customer, starts_on;
