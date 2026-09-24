-- Сколько РАЗНЫХ товаров купил каждый клиент
SELECT o.customer,
       count(i.product_id) AS products
FROM orders o
JOIN order_items i ON i.order_id = o.id
GROUP BY o.customer
ORDER BY o.customer;
