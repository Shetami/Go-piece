-- Закупка базовых деталей на 5 велосипедов (parts.id = 1).
-- Колонки: name, unit, qty = round(…, 3), cost = round(qty * unit_cost, 2). Порядок: cost DESC, name.
SELECT p.name, p.unit, b.qty * 5 AS qty, round(b.qty * 5 * p.unit_cost, 2) AS cost
FROM bom b
JOIN parts p ON p.id = b.child_id
WHERE b.parent_id = 1
ORDER BY cost DESC, p.name;
