-- Доля возвратов по категориям.
-- Колонки: category, sold, refunded, refund_pct (round(…, 1), NULL если продаж нет).
-- Порядок: refund_pct по убыванию (NULL в конце), затем category.
SELECT category,
       sum(qty) AS sold,
       sum(refunded_qty) AS refunded,
       round(sum(refunded_qty) * 100 / sum(qty), 1) AS refund_pct
FROM order_items
WHERE qty > 0
GROUP BY category
ORDER BY refund_pct DESC, category;
