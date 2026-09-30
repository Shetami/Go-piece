-- Доля отмен по курьерам (от 5 доставок) и отклонение от средней по компании.
-- Колонки: courier, deliveries, cancelled, cancel_pct (round 1), vs_avg (п. п., round 1).
-- Порядок: cancel_pct по убыванию, затем courier.
SELECT courier,
       count(*) AS deliveries,
       count(*) FILTER (WHERE status = 'cancelled') AS cancelled,
       round(count(*) FILTER (WHERE status = 'cancelled') / count(*) * 100, 1) AS cancel_pct,
       round(100.0 * count(*) FILTER (WHERE status = 'cancelled') / count(*)
             - 100.0 * sum(count(*) FILTER (WHERE status = 'cancelled')) OVER ()
                     / sum(count(*)) OVER (), 1) AS vs_avg
FROM deliveries
GROUP BY courier
HAVING count(*) >= 5
ORDER BY cancel_pct DESC, courier;
