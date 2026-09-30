-- Уникальные пользователи по дням недели 7–13 октября 2024 и за всю неделю.
-- Колонки: day ('YYYY-MM-DD', последняя строка 'неделя'), users, events.
-- Порядок: дни по возрастанию, 'неделя' — последней.
WITH daily AS (
  SELECT to_char(happened_at, 'YYYY-MM-DD') AS day,
         count(DISTINCT user_id) AS users,
         count(*) AS events
  FROM events
  WHERE happened_at BETWEEN '2024-10-07' AND '2024-10-13'
  GROUP BY 1
)
SELECT day, users, events FROM daily
UNION ALL
SELECT 'неделя', sum(users), sum(events) FROM daily
ORDER BY 1;
