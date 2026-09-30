-- Остаток товара на складе в момент запроса аудитора (включительно).
-- Колонки: check_id, qty. Порядок: по check_id.
SELECT s.id AS check_id, sum(mv.delta) AS qty
FROM stock_checks s
JOIN movements mv ON mv.sku = s.sku AND mv.moved_at < s.at
GROUP BY s.id
ORDER BY s.id;
