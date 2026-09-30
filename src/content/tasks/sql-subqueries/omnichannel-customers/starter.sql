-- По каждому городу: сколько клиентов всего, сколько покупали и онлайн, и в магазине, их доля в %.
-- Колонки: city, customers, omni, omni_pct = round(…, 1). Порядок: city.
SELECT c.city,
       count(*) AS customers,
       count(*) FILTER (WHERE o.id IS NOT NULL AND s.id IS NOT NULL) AS omni,
       round(100.0 * count(*) FILTER (WHERE o.id IS NOT NULL AND s.id IS NOT NULL) / count(*), 1) AS omni_pct
FROM customers c
LEFT JOIN purchases o ON o.customer_id = c.id AND o.channel = 'online'
LEFT JOIN purchases s ON s.customer_id = c.id AND s.channel = 'store'
GROUP BY c.city
ORDER BY c.city;
