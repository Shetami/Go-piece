-- Залежавшиеся товары на 2024-06-30: остаток по всем складам > 0, а последней продажи (не возврата)
-- не было с 2024-06-01 00:00 (или продаж не было вовсе).
-- Колонки: name, stock, last_sale (дата), days_idle (2024-06-30 - last_sale). Порядок: name.
SELECT p.name,
       sum(st.qty)                                  AS stock,
       max(s.sold_at)::date                         AS last_sale,
       date '2024-06-30' - max(s.sold_at)::date     AS days_idle
FROM products p
JOIN stock st ON st.product_id = p.id
JOIN sales s  ON s.product_id = p.id
GROUP BY p.name
HAVING sum(st.qty) > 0
   AND max(s.sold_at) < '2024-06-01'
ORDER BY p.name;
