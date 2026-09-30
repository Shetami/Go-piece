-- Самая длинная серия дней подряд с тренировками, по местному времени.
-- Колонки: user_name, longest_streak, streak_start. Порядок: longest_streak по убыванию, затем user_name.
WITH islands AS (
  SELECT user_id, done_at::date AS day,
         done_at::date - row_number() OVER (PARTITION BY user_id ORDER BY done_at)::int AS grp
  FROM workouts
)
SELECT u.name AS user_name,
       max(cnt) AS longest_streak,
       min(started_on) AS streak_start
FROM users u
JOIN (SELECT user_id, grp, count(*) AS cnt, min(day) AS started_on
      FROM islands GROUP BY user_id, grp) s ON s.user_id = u.id
GROUP BY u.name
ORDER BY longest_streak DESC, user_name;
