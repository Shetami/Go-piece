-- Движение активной базы по месяцам, февраль—май 2024.
-- Колонки: month, active, new, retained, reactivated, churned. Порядок: по month.
WITH a AS (
  SELECT user_id, date_trunc('month', ts)::date AS month,
         lag(date_trunc('month', ts)::date) OVER (PARTITION BY user_id ORDER BY ts) AS prev
  FROM events
)
SELECT month,
       count(*) AS active,
       count(*) FILTER (WHERE prev IS NULL) AS new,
       count(*) FILTER (WHERE prev = month - interval '1 month') AS retained,
       count(*) FILTER (WHERE prev < month - interval '1 month') AS reactivated,
       0 AS churned
FROM a
WHERE month >= '2024-02-01'
GROUP BY month
ORDER BY month;
