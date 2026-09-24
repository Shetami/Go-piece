WITH days AS (
  SELECT DISTINCT user_name, day
  FROM logins
),
grouped AS (
  SELECT user_name, day,
         day - row_number() OVER (PARTITION BY user_name ORDER BY day)::int AS grp
  FROM days
),
streaks AS (
  SELECT user_name, grp, count(*) AS len
  FROM grouped
  GROUP BY user_name, grp
)
SELECT user_name, max(len) AS longest_streak
FROM streaks
GROUP BY user_name
ORDER BY user_name;
