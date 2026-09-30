-- Активные пользователи по дням с 1 по 7 июня 2024.
-- Колонки: day (date), active_users, pro_users. Порядок: по day.
SELECT d.day::date AS day,
       count(*) AS active_users,
       count(*) FILTER (WHERE s.plan = 'pro') AS pro_users
FROM generate_series(date '2024-06-01', date '2024-06-07', interval '1 day') AS d(day)
JOIN subscriptions s ON d.day BETWEEN s.started_at AND s.ended_at
GROUP BY d.day
ORDER BY d.day;
