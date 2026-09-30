WITH items AS (
  SELECT o.seller_id, i.id, i.price * i.qty AS gross,
         o.delivered_at::date <= date '2024-11-15' - 14 AS released,
         (SELECT r.pct
          FROM commission_rates r
          WHERE r.category_id = i.category_id
            AND r.valid_from <= o.created_on
          ORDER BY r.valid_from DESC
          LIMIT 1) AS pct
  FROM order_items i
  JOIN orders o ON o.id = i.order_id
  WHERE o.delivered_at IS NOT NULL
    AND NOT EXISTS (SELECT 1 FROM returns r    WHERE r.order_item_id = i.id)
    AND NOT EXISTS (SELECT 1 FROM paid_items p WHERE p.order_item_id = i.id)
),
per_seller AS (
  SELECT seller_id,
         count(*) FILTER (WHERE released)                   AS items,
         sum(gross) FILTER (WHERE released)                 AS gross,
         sum(round(gross * pct / 100, 2)) FILTER (WHERE released) AS commission,
         sum(gross) FILTER (WHERE NOT released)             AS on_hold
  FROM items
  GROUP BY seller_id
)
SELECT s.name AS seller,
       coalesce(p.items, 0)                            AS items,
       coalesce(p.gross, 0)                            AS gross,
       coalesce(p.commission, 0)                       AS commission,
       coalesce(p.gross, 0) - coalesce(p.commission, 0) AS payout,
       coalesce(p.on_hold, 0)                          AS on_hold
FROM sellers s
LEFT JOIN per_seller p ON p.seller_id = s.id
ORDER BY payout DESC, seller;
