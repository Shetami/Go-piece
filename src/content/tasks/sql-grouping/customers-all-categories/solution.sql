SELECT o.customer,
       count(DISTINCT o.id) AS orders
FROM orders o
JOIN order_items i ON i.order_id = o.id
JOIN products p    ON p.id = i.product_id
JOIN categories c  ON c.id = p.category_id AND c.is_active
WHERE o.status = 'paid'
GROUP BY o.customer
HAVING count(DISTINCT c.id) = (SELECT count(*) FROM categories WHERE is_active)
ORDER BY o.customer;
