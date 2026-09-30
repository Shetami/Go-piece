-- Пары товаров, которые покупали вместе хотя бы в двух оплаченных заказах.
-- Колонки: product_a, product_b (id первого меньше id второго), orders_together.
-- Порядок: orders_together по убыванию, затем id первого товара, затем id второго.
SELECT pa.name AS product_a, pb.name AS product_b, count(*) AS orders_together
FROM order_items a
JOIN order_items b ON b.order_id = a.order_id AND a.product_id <> b.product_id
JOIN products pa ON pa.id = a.product_id
JOIN products pb ON pb.id = b.product_id
GROUP BY a.product_id, b.product_id, pa.name, pb.name
HAVING count(*) >= 2
ORDER BY orders_together DESC, a.product_id, b.product_id;
