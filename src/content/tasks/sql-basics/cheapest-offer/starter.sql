-- Лучшее предложение по каждому товару.
-- Колонки: product, seller, price (NULL, если купить негде).
-- Порядок: по product.
SELECT p.name AS product, s.name AS seller, min(o.price) AS price
FROM products p
JOIN offers o ON o.product_id = p.id
JOIN sellers s ON s.id = o.seller_id
GROUP BY p.name, s.name
ORDER BY p.name;
