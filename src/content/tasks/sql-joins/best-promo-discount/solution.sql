WITH item_discount AS (
  SELECT i.order_id,
         i.qty * i.price                     AS gross,
         coalesce(max(pr.discount_pct), 0)   AS pct
  FROM order_items i
  JOIN orders o   ON o.id = i.order_id
  JOIN products p ON p.id = i.product_id
  LEFT JOIN promotions pr
    ON (pr.category_id = p.category_id OR pr.category_id IS NULL)
   AND pr.starts_at <= o.created_at
   AND (pr.ends_at IS NULL OR o.created_at < pr.ends_at)
  GROUP BY i.id, i.order_id, i.qty, i.price
)
SELECT order_id,
       sum(gross)                                AS gross,
       round(sum(gross * pct / 100), 2)          AS discount,
       round(sum(gross * (1 - pct / 100)), 2)    AS net
FROM item_discount
GROUP BY order_id
ORDER BY order_id;
