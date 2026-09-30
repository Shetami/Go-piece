-- Средняя цена закупки по товарам, взвешенная по количеству.
-- Колонки: sku, units, avg_price (round 2), min_price, max_price.
-- Только товары с units >= 10. Порядок: avg_price по убыванию, затем sku.
SELECT sku,
       sum(qty) AS units,
       round(avg(unit_price), 2) AS avg_price,
       min(unit_price) AS min_price,
       max(unit_price) AS max_price
FROM purchases
GROUP BY sku
HAVING sum(qty) >= 10
ORDER BY avg_price DESC, sku;
