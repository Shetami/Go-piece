SELECT c.city,
       count(*) AS customers,
       count(*) FILTER (
         WHERE EXISTS (SELECT 1 FROM purchases p WHERE p.customer_id = c.id AND p.channel = 'online')
           AND EXISTS (SELECT 1 FROM purchases p WHERE p.customer_id = c.id AND p.channel = 'store')
       ) AS omni,
       round(100.0 * count(*) FILTER (
         WHERE EXISTS (SELECT 1 FROM purchases p WHERE p.customer_id = c.id AND p.channel = 'online')
           AND EXISTS (SELECT 1 FROM purchases p WHERE p.customer_id = c.id AND p.channel = 'store')
       ) / count(*), 1) AS omni_pct
FROM customers c
GROUP BY c.city
ORDER BY c.city;
