-- Движение товара по складу.
-- Колонки: sku, received, sold (только оплаченные заказы), returned,
--          on_hand = received - sold + returned. Порядок: по sku.
SELECT p.sku,
       sum(r.qty)  AS received,
       sum(i.qty)  AS sold,
       sum(rt.qty) AS returned,
       sum(r.qty) - sum(i.qty) + sum(rt.qty) AS on_hand
FROM products p
JOIN receipts r ON r.product_id = p.id
JOIN order_items i ON i.product_id = p.id
LEFT JOIN returns rt ON rt.product_id = p.id
GROUP BY p.sku
ORDER BY p.sku;
