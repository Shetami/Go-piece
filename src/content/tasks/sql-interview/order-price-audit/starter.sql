-- Позиции оплаченных заказов, где цена не совпала с ожидаемой.
-- Колонки: order_id, product, qty, charged, expected, diff_total. Порядок: order_id, product.
SELECT i.order_id,
       p.name AS product,
       i.qty,
       i.unit_price AS charged,
       ph.price AS expected,
       (i.unit_price - ph.price) * i.qty AS diff_total
FROM order_items i
JOIN orders o         ON o.id = i.order_id
JOIN products p       ON p.id = i.product_id
JOIN price_history ph ON ph.product_id = i.product_id
                     AND o.created_at BETWEEN ph.valid_from AND ph.valid_to
WHERE i.unit_price <> ph.price
ORDER BY i.order_id, product;
