-- Выручка и число товаров по каждой категории вместе со всеми её подкатегориями.
-- Колонки: name, revenue (0, если продаж нет), products. Порядок: revenue DESC, name.
SELECT c.name,
       sum(s.qty * s.price) AS revenue,
       count(p.id)          AS products
FROM categories c
JOIN products p    ON p.category_id = c.id
LEFT JOIN sales s  ON s.product_id = p.id
GROUP BY c.id, c.name
ORDER BY revenue DESC, c.name;
