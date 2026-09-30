-- Отправления по неделям (понедельник) и статусам, недели с 1 по 29 апреля 2024.
-- Колонки: week (date), total, delivered, in_transit, lost, other.
-- Порядок: по week.
SELECT date_trunc('week', created_at::timestamp)::date AS week,
       count(*) AS total,
       count(*) FILTER (WHERE status = 'delivered')  AS delivered,
       count(*) FILTER (WHERE status = 'in_transit') AS in_transit,
       count(*) FILTER (WHERE status = 'lost')       AS lost,
       count(*) FILTER (WHERE status NOT IN ('delivered', 'in_transit', 'lost')) AS other
FROM shipments
WHERE created_at BETWEEN '2024-04-01' AND '2024-04-29'
GROUP BY 1
ORDER BY 1;
