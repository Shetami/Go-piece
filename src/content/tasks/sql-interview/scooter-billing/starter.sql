-- Счёт за аренду самокатов по пользователям.
-- Колонки: user_name, rides, minutes, amount. Порядок: amount по убыванию, затем user_name.
SELECT u.name AS user_name,
       count(*) AS rides,
       sum(round(extract(epoch FROM r.ended_at - r.started_at) / 60)) AS minutes,
       sum(round(extract(epoch FROM r.ended_at - r.started_at) / 60) * t.price_per_min + 50) AS amount
FROM users u
JOIN rides r   ON r.user_id = u.id
JOIN tariffs t ON r.started_at::time >= t.from_time AND r.started_at::time < t.to_time
GROUP BY u.name
ORDER BY amount DESC, user_name;
