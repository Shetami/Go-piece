-- Отчёт по дням с 2024-02-26 по 2024-03-03 включительно, дни без продаж — нулями.
-- Колонки: day, orders, revenue, running (нарастающая выручка с начала периода). Порядок: day.
SELECT created_at::date AS day,
       count(*)    AS orders,
       sum(amount) AS revenue,
       sum(sum(amount)) OVER (ORDER BY created_at::date) AS running
FROM orders
WHERE status = 'paid'
  AND created_at BETWEEN '2024-02-26' AND '2024-03-03'
GROUP BY created_at::date
ORDER BY day;
