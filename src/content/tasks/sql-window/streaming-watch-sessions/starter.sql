-- Сессии просмотра: heartbeat-ы пользователя в одной сессии, пока пауза между соседними
-- не больше 5 минут. Минуты сессии = число разных heartbeat-ов (повторы с тем же ts не считаются).
-- Колонки: user_name, sessions, total_minutes, longest_minutes. Все пользователи, без просмотров — нули.
-- Порядок: user_name.
SELECT u.name AS user_name,
       count(DISTINCT h.ts::date) AS sessions,
       count(h.ts)                AS total_minutes,
       count(h.ts)                AS longest_minutes
FROM users u
JOIN heartbeats h ON h.user_id = u.id
GROUP BY u.name
ORDER BY u.name;
