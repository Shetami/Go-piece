-- Два самых продаваемых товара каждой категории за май 2024 (только доставленные заказы).
-- Колонки: category, rank (1 или 2), product, units.
-- Категория без продаж — одна строка с NULL в rank, product, units.
-- Порядок: по category, затем по rank.
SELECT c.name AS category, 1 AS rank, p.name AS product, sum(i.qty) AS units
FROM categories c
JOIN products p    ON p.category_id = c.id
JOIN order_items i ON i.product_id = p.id
GROUP BY c.name, p.name
ORDER BY c.name, units DESC;
