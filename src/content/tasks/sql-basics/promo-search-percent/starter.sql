-- Товары акции «50%», кроме уценённых.
-- Колонки: id, name, promo_price (половина цены, round(…, 2)).
-- Порядок: по id.
SELECT id, name, price / 2 AS promo_price
FROM products
WHERE name LIKE '%50%%'
  AND description NOT LIKE '%уценк%'
ORDER BY id;
