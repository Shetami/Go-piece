WITH marked AS (
  SELECT user_id, ts,
         CASE WHEN ts - lag(ts) OVER w <= INTERVAL '5 minutes' THEN 0 ELSE 1 END AS is_new
  FROM heartbeats
  WINDOW w AS (PARTITION BY user_id ORDER BY ts)
),
numbered AS (
  SELECT user_id, ts,
         sum(is_new) OVER (PARTITION BY user_id ORDER BY ts
                           ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW) AS session_no
  FROM marked
),
sessions AS (
  SELECT user_id, session_no, count(DISTINCT ts) AS minutes
  FROM numbered
  GROUP BY user_id, session_no
)
SELECT u.name AS user_name,
       count(s.session_no)          AS sessions,
       coalesce(sum(s.minutes), 0)  AS total_minutes,
       coalesce(max(s.minutes), 0)  AS longest_minutes
FROM users u
LEFT JOIN sessions s ON s.user_id = u.id
GROUP BY u.id, u.name
ORDER BY u.name;
