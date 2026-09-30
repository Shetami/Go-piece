SELECT s.id AS check_id,
       coalesce(c.qty, 0) + coalesce(m.delta, 0) AS qty
FROM stock_checks s
LEFT JOIN LATERAL (
  SELECT ic.qty, ic.counted_at
  FROM inventory_counts ic
  WHERE ic.warehouse_id = s.warehouse_id
    AND ic.sku = s.sku
    AND ic.counted_at <= s.at
  ORDER BY ic.counted_at DESC
  LIMIT 1
) c ON true
LEFT JOIN LATERAL (
  SELECT sum(mv.delta) AS delta
  FROM movements mv
  WHERE mv.warehouse_id = s.warehouse_id
    AND mv.sku = s.sku
    AND mv.moved_at <= s.at
    AND (c.counted_at IS NULL OR mv.moved_at > c.counted_at)
) m ON true
ORDER BY s.id;
