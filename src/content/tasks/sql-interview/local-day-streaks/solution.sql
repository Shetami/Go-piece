WITH days AS (
  -- одна строка на локальный день пользователя
  SELECT DISTINCT w.user_id, (w.done_at AT TIME ZONE u.tz)::date AS day
  FROM workouts w
  JOIN users u ON u.id = w.user_id
),
islands AS (
  SELECT user_id, day,
         day - row_number() OVER (PARTITION BY user_id ORDER BY day)::int AS grp
  FROM days
),
streaks AS (
  SELECT user_id, min(day) AS started_on, count(*) AS len
  FROM islands
  GROUP BY user_id, grp
),
best AS (
  SELECT DISTINCT ON (user_id) user_id, len, started_on
  FROM streaks
  ORDER BY user_id, len DESC, started_on
)
SELECT u.name AS user_name,
       coalesce(b.len, 0) AS longest_streak,
       b.started_on AS streak_start
FROM users u
LEFT JOIN best b ON b.user_id = u.id
ORDER BY longest_streak DESC, user_name;
