-- Что дозаказать: товары, у которых остаток + то, что в пути, ниже минимума.
-- Колонки: warehouse, sku, on_hand, in_transit, order_qty (target − on_hand − in_transit).
-- Порядок: по warehouse, затем по sku.
SELECT w.name AS warehouse, s.sku, sum(s.qty) AS on_hand,
       coalesce(sum(po.qty), 0) AS in_transit,
       max(r.target_qty) - sum(s.qty) - coalesce(sum(po.qty), 0) AS order_qty
FROM stock s
JOIN warehouses w ON w.id = s.warehouse_id
JOIN reorder_rules r ON r.sku = s.sku
LEFT JOIN purchase_orders po ON po.warehouse_id = s.warehouse_id AND po.sku = s.sku
GROUP BY w.name, s.sku
HAVING sum(s.qty) + coalesce(sum(po.qty), 0) < max(r.min_qty)
ORDER BY warehouse, sku;
