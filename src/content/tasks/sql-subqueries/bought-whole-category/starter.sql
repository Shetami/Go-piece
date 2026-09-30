-- Пары (категория, клиент): клиент купил и не вернул каждый активный товар категории.
-- Колонки: category, name. Порядок: category, name.
SELECT p.category, c.name
FROM purchases x
JOIN products p  ON p.id = x.product_id
JOIN customers c ON c.id = x.customer_id
GROUP BY p.category, c.name
HAVING count(*) >= (SELECT count(*) FROM products p2 WHERE p2.category = p.category)
ORDER BY p.category, c.name;
