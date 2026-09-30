-- Гистограмма оплаченных заказов по сумме: корзины по 1000 от 0 до 5000 и 5000+.
-- Колонки: lo, hi (не включая; NULL у последней), orders, share_pct (round 1).
-- Порядок: по lo.
SELECT floor(amount / 1000) * 1000     AS lo,
       floor(amount / 1000) * 1000 + 1000 AS hi,
       count(*) AS orders,
       round(100.0 * count(*) / (SELECT count(*) FROM orders), 1) AS share_pct
FROM orders
WHERE status = 'paid'
GROUP BY 1
ORDER BY 1;
