WITH rules AS (
  SELECT DISTINCT ON (warehouse_id, sku) warehouse_id, sku, min_qty, target_qty
  FROM reorder_rules
  ORDER BY warehouse_id, sku, updated_at DESC
),
on_hand AS (
  SELECT warehouse_id, sku, sum(qty) AS qty
  FROM stock
  GROUP BY warehouse_id, sku
),
transit AS (
  SELECT warehouse_id, sku, sum(qty) AS qty
  FROM purchase_orders
  WHERE status = 'open'
  GROUP BY warehouse_id, sku
),
grid AS (
  SELECT w.id AS warehouse_id, w.name, p.sku,
         coalesce(own.min_qty, dflt.min_qty)       AS min_qty,
         coalesce(own.target_qty, dflt.target_qty) AS target_qty,
         coalesce(h.qty, 0) AS on_hand,
         coalesce(t.qty, 0) AS in_transit
  FROM warehouses w
  CROSS JOIN products p
  LEFT JOIN rules own  ON own.warehouse_id = w.id AND own.sku = p.sku
  LEFT JOIN rules dflt ON dflt.warehouse_id IS NULL AND dflt.sku = p.sku
  LEFT JOIN on_hand h  ON h.warehouse_id = w.id AND h.sku = p.sku
  LEFT JOIN transit t  ON t.warehouse_id = w.id AND t.sku = p.sku
  WHERE w.is_active AND p.is_active
)
SELECT name AS warehouse, sku, on_hand, in_transit,
       target_qty - on_hand - in_transit AS order_qty
FROM grid
WHERE on_hand + in_transit < min_qty
ORDER BY name, sku;
