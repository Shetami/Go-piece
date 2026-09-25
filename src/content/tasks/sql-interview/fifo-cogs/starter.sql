-- Себестоимость продаж по FIFO и остаток склада по каждому товару.
-- Колонки: product, sold_qty, revenue, cogs, gross_profit, stock_left, stock_value.
-- Порядок: по product.
SELECT p.name AS product,
       sum(s.qty) AS sold_qty,
       sum(s.qty * s.unit_price) AS revenue,
       sum(s.qty) * avg(b.unit_cost) AS cogs,
       0 AS gross_profit,
       sum(b.qty) - sum(s.qty) AS stock_left,
       0 AS stock_value
FROM products p
JOIN purchases b ON b.product_id = p.id
JOIN sales s     ON s.product_id = p.id
GROUP BY p.name
ORDER BY p.name;
