SELECT id, name, round(price / 2.0, 2) AS promo_price   -- 2.0: не целочисленное деление
FROM products
WHERE name LIKE '%50\%%'                                -- \% — буквальный знак процента
  AND (description IS NULL OR description NOT ILIKE '%уценк%')
ORDER BY id;
