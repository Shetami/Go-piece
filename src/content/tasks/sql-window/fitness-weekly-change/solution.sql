WITH weeks AS (
  SELECT generate_series(
           date_trunc('week', min(logged_at)),
           date_trunc('week', max(logged_at)),
           INTERVAL '1 week'
         )::date AS week_start
  FROM activity
),
users AS (
  SELECT DISTINCT user_name FROM activity
),
weekly AS (
  SELECT user_name, date_trunc('week', logged_at)::date AS week_start, sum(steps) AS steps
  FROM activity
  GROUP BY 1, 2
),
grid AS (
  SELECT u.user_name, w.week_start, coalesce(x.steps, 0) AS steps
  FROM users u
  CROSS JOIN weeks w
  LEFT JOIN weekly x ON x.user_name = u.user_name AND x.week_start = w.week_start
)
SELECT user_name, week_start, steps,
       round(100.0 * (steps - lag(steps) OVER w) / nullif(lag(steps) OVER w, 0), 1) AS change_pct
FROM grid
WINDOW w AS (PARTITION BY user_name ORDER BY week_start)
ORDER BY user_name, week_start;
