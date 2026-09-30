-- Скидка по заказам: к каждой позиции применяется одна, самая выгодная из действующих акций.
-- Колонки: order_id, gross, discount (round(…, 2)), net (round(…, 2)). Порядок: по order_id.
SELECT o.id AS order_id, sum(i.qty * i.price) AS gross,
       round(sum(i.qty * i.price * pr.discount_pct / 100), 2) AS discount,
       round(sum(i.qty * i.price * (1 - pr.discount_pct / 100)), 2) AS net
FROM orders o
JOIN order_items i ON i.order_id = o.id
JOIN products p    ON p.id = i.product_id
JOIN promotions pr ON pr.category_id = p.category_id
                  AND o.created_at BETWEEN pr.starts_at AND pr.ends_at
GROUP BY o.id
ORDER BY o.id;
