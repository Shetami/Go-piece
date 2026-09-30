-- Остаток каждого товара на конец дня, 1–5 августа 2024.
-- Колонки: sku, day (date), balance. Порядок: по sku, затем по day.
SELECT sku, moved_at::date AS day,
       sum(sum(qty)) OVER (PARTITION BY sku ORDER BY moved_at::date) AS balance
FROM stock_moves
WHERE moved_at::date BETWEEN '2024-08-01' AND '2024-08-05'
GROUP BY sku, moved_at::date
ORDER BY sku, day;
