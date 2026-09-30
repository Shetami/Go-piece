SELECT coalesce(nullif(lower(trim(category)), ''), 'без категории') AS category,
       sum(qty)                        AS sold,
       coalesce(sum(refunded_qty), 0)  AS refunded,
       round(100.0 * coalesce(sum(refunded_qty), 0) / nullif(sum(qty), 0), 1) AS refund_pct
FROM order_items
GROUP BY 1
ORDER BY refund_pct DESC NULLS LAST, category;
