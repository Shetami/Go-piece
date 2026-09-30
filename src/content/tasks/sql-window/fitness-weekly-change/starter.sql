-- Шаги по неделям (с понедельника) для каждого пользователя — по всем неделям от первой
-- до последней в данных, пустые недели с нулём; изменение к прошлой неделе в процентах, round 1.
-- NULL, если прошлой недели нет или в ней 0 шагов.
-- Колонки: user_name, week_start (date), steps, change_pct. Порядок: user_name, week_start.
WITH weekly AS (
  SELECT user_name, date_trunc('week', logged_at)::date AS week_start, sum(steps) AS steps
  FROM activity
  GROUP BY 1, 2
)
SELECT user_name, week_start, steps,
       round(100 * (steps - lag(steps) OVER w) / lag(steps) OVER w, 1) AS change_pct
FROM weekly
WINDOW w AS (PARTITION BY user_name ORDER BY week_start)
ORDER BY user_name, week_start;
