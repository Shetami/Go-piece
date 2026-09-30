-- Выручка по тарифам за каждый месяц и её доля в выручке месяца, %.
-- Выручка = paid минус refunded; failed не считаются. Тариф попадает в месяц, если в нём
-- был хотя бы один paid или refunded платёж.
-- Колонки: month ('YYYY-MM'), plan, revenue, share_pct (round 1). Порядок: month, revenue по убыванию, plan.
SELECT to_char(paid_at, 'YYYY-MM') AS month, plan, sum(amount) AS revenue,
       round(100 * sum(amount) / (SELECT sum(amount) FROM payments), 1) AS share_pct
FROM payments
GROUP BY 1, 2
ORDER BY month, revenue DESC, plan;
