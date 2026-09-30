SELECT p.sku,
       p.price,
       (SELECT min(c.price) FROM competitor_prices c WHERE c.sku = p.sku) AS best_competitor,
       (SELECT min(c.price) FROM competitor_prices c WHERE c.sku = p.sku) - p.price AS gap
FROM products p
WHERE EXISTS (
        SELECT 1 FROM competitor_prices c
        WHERE c.sku = p.sku AND c.price IS NOT NULL
      )
  AND p.price < ALL (
        SELECT c.price FROM competitor_prices c
        WHERE c.sku = p.sku AND c.price IS NOT NULL
      )
ORDER BY gap DESC, p.sku;
