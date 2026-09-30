-- Какие новые заказы можно собрать из остатков всех складов.
-- Колонки: order_id, items (разных товаров), missing (товаров не хватает),
--          can_ship (boolean). Порядок: по order_id.
SELECT i.order_id,
       count(*) AS items,
       count(*) FILTER (WHERE s.qty < i.qty) AS missing,
       bool_and(s.qty >= i.qty) AS can_ship
FROM order_items i
JOIN orders o ON o.id = i.order_id
LEFT JOIN stock s ON s.product_id = i.product_id
WHERE o.status = 'new'
GROUP BY i.order_id
ORDER BY i.order_id;
