SELECT p.name AS product, b.seller, b.price
FROM products p
LEFT JOIN (
  SELECT DISTINCT ON (o.product_id)
         o.product_id, s.name AS seller, o.price
  FROM offers o
  JOIN sellers s ON s.id = o.seller_id
  WHERE o.stock > 0                               -- NULL > 0 — не true, неизвестный остаток отсекается
  ORDER BY o.product_id, o.price, s.rating DESC NULLS LAST, s.id
) b ON b.product_id = p.id
ORDER BY p.name;
