-- Товары, которые у нас строго дешевле всех известных цен конкурентов (хотя бы одна цена известна).
-- Колонки: sku, price, best_competitor, gap = best_competitor - price. Порядок: gap DESC, sku.
SELECT p.sku,
       p.price,
       (SELECT min(c.price) FROM competitor_prices c WHERE c.sku = p.sku) AS best_competitor,
       (SELECT min(c.price) FROM competitor_prices c WHERE c.sku = p.sku) - p.price AS gap
FROM products p
WHERE p.price < ALL (SELECT c.price FROM competitor_prices c WHERE c.sku = p.sku)
ORDER BY gap DESC, p.sku;
