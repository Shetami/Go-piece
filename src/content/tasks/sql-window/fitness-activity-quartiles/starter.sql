-- Квартили активности: ntile(4) по сумме минут (больше — выше), при равенстве — по имени.
-- pct_rank — percent_rank по возрастанию суммы, round 2. Все пользователи, без тренировок — 0 минут.
-- Колонки: name, total_minutes, quartile, pct_rank. Порядок: quartile, total_minutes по убыванию, name.
SELECT u.name, sum(w.minutes) AS total_minutes,
       ntile(4) OVER (ORDER BY sum(w.minutes) DESC) AS quartile,
       round((rank() OVER (ORDER BY sum(w.minutes)))::numeric / count(*) OVER (), 2) AS pct_rank
FROM users u
JOIN workouts w ON w.user_id = u.id
GROUP BY u.name
ORDER BY quartile, total_minutes DESC, name;
