WITH totals AS (
  SELECT u.name, coalesce(sum(w.minutes), 0) AS total_minutes
  FROM users u
  LEFT JOIN workouts w ON w.user_id = u.id
  GROUP BY u.id, u.name
)
SELECT name, total_minutes,
       ntile(4) OVER (ORDER BY total_minutes DESC, name) AS quartile,
       round(percent_rank() OVER (ORDER BY total_minutes)::numeric, 2) AS pct_rank
FROM totals
ORDER BY quartile, total_minutes DESC, name;
