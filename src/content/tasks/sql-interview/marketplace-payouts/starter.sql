-- Выплата продавцам на 2024-11-15.
-- Колонки: seller, items, gross, commission, payout, on_hold. Порядок: payout по убыванию, затем seller.
SELECT s.name AS seller,
       count(i.id) AS items,
       sum(i.price * i.qty) AS gross,
       sum(i.price * i.qty * r.pct / 100) AS commission,
       sum(i.price * i.qty * (1 - r.pct / 100)) AS payout,
       0 AS on_hold
FROM sellers s
JOIN orders o           ON o.seller_id = s.id
JOIN order_items i      ON i.order_id = o.id
JOIN commission_rates r ON r.category_id = i.category_id
WHERE o.delivered_at < '2024-11-01'
GROUP BY s.name
ORDER BY payout DESC, seller;
