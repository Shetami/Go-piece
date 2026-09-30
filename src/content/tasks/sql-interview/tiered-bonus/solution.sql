WITH returned AS (
  SELECT sale_id, sum(amount) AS amount
  FROM returns
  GROUP BY sale_id
),
net AS (
  SELECT m.id, m.name,
         coalesce(sum(s.amount - coalesce(r.amount, 0)), 0) AS net_sales
  FROM managers m
  LEFT JOIN sales s    ON s.manager_id = m.id
                      AND s.sold_on >= date '2024-04-01'
                      AND s.sold_on <  date '2024-07-01'
  LEFT JOIN returned r ON r.sale_id = s.id
  GROUP BY m.id, m.name
)
SELECT n.name AS manager,
       n.net_sales,
       round(coalesce(sum(
         (least(n.net_sales, coalesce(t.to_amount, n.net_sales)) - t.from_amount) * t.pct / 100
       ), 0), 2) AS bonus,
       max(t.pct) AS top_tier_pct
FROM net n
LEFT JOIN bonus_tiers t ON n.net_sales > t.from_amount
GROUP BY n.id, n.name, n.net_sales
ORDER BY bonus DESC, manager;
