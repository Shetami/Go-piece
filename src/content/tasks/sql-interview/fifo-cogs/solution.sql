WITH batches AS (
  SELECT product_id, qty, unit_cost,
         sum(qty) OVER w - qty AS from_qty, -- сколько единиц лежало на складе до этой партии
         sum(qty) OVER w       AS to_qty
  FROM purchases
  WINDOW w AS (PARTITION BY product_id ORDER BY bought_at, id)
),
sold AS (
  SELECT product_id, sum(qty) AS qty, sum(qty * unit_price) AS revenue
  FROM sales
  GROUP BY product_id
),
consumed AS (
  SELECT b.product_id, b.qty, b.unit_cost,
         greatest(0, least(b.to_qty, coalesce(s.qty, 0)) - b.from_qty) AS used
  FROM batches b
  LEFT JOIN sold s ON s.product_id = b.product_id
),
cost AS (
  SELECT product_id,
         sum(used * unit_cost)          AS cogs,
         sum(qty - used)                AS stock_left,
         sum((qty - used) * unit_cost)  AS stock_value
  FROM consumed
  GROUP BY product_id
)
SELECT p.name AS product,
       coalesce(s.qty, 0)                          AS sold_qty,
       coalesce(s.revenue, 0)                      AS revenue,
       coalesce(c.cogs, 0)                         AS cogs,
       coalesce(s.revenue, 0) - coalesce(c.cogs, 0) AS gross_profit,
       coalesce(c.stock_left, 0)                   AS stock_left,
       coalesce(c.stock_value, 0)                  AS stock_value
FROM products p
LEFT JOIN sold s ON s.product_id = p.id
LEFT JOIN cost c ON c.product_id = p.id
ORDER BY p.name;
