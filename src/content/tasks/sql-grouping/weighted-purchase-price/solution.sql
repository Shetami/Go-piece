SELECT sku,
       sum(qty)                                   AS units,
       round(sum(qty * unit_price) / sum(qty), 2) AS avg_price,
       min(unit_price)                            AS min_price,
       max(unit_price)                            AS max_price
FROM purchases
WHERE qty > 0 AND unit_price IS NOT NULL
GROUP BY sku
HAVING sum(qty) >= 10
ORDER BY avg_price DESC, sku;
