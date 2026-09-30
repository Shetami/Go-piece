WITH RECURSIVE keys AS (
  -- нормализованные ключи; пустые значения ключом не являются
  SELECT id, 'e:' || lower(trim(email)) AS k
  FROM customers
  WHERE nullif(trim(email), '') IS NOT NULL
  UNION ALL
  SELECT id, 'p:' || regexp_replace(phone, '\D', '', 'g')
  FROM customers
  WHERE regexp_replace(coalesce(phone, ''), '\D', '', 'g') <> ''
),
edges AS (
  SELECT DISTINCT a.id AS src, b.id AS dst
  FROM keys a
  JOIN keys b ON b.k = a.k AND b.id <> a.id
),
reach AS (
  SELECT id AS start_id, id AS node_id FROM customers
  UNION
  SELECT r.start_id, e.dst
  FROM reach r
  JOIN edges e ON e.src = r.node_id
),
components AS (
  SELECT start_id AS id, min(node_id) AS master_id
  FROM reach
  GROUP BY start_id
),
spend AS (
  SELECT customer_id, count(*) AS orders, sum(amount) AS revenue
  FROM orders
  GROUP BY customer_id
)
SELECT c.master_id,
       m.name AS master_name,
       count(*) AS accounts,
       coalesce(sum(s.orders), 0)  AS orders,
       coalesce(sum(s.revenue), 0) AS revenue
FROM components c
JOIN customers m  ON m.id = c.master_id
LEFT JOIN spend s ON s.customer_id = c.id
GROUP BY c.master_id, m.name
HAVING count(*) > 1
ORDER BY c.master_id;
