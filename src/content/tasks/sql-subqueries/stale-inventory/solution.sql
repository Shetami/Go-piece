WITH on_hand AS (
  SELECT product_id, sum(qty) AS stock
  FROM stock
  GROUP BY product_id
  HAVING sum(qty) > 0
),
report AS (
  SELECT p.name,
         h.stock,
         (SELECT max(s.sold_at) FROM sales s
          WHERE s.product_id = p.id AND s.kind = 'sale') AS last_sale
  FROM products p
  JOIN on_hand h ON h.product_id = p.id
)
SELECT name,
       stock,
       last_sale::date                    AS last_sale,
       date '2024-06-30' - last_sale::date AS days_idle
FROM report
WHERE last_sale IS NULL
   OR last_sale < timestamp '2024-06-01'
ORDER BY name;
