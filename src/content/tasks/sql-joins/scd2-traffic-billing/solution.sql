WITH priced AS (
  SELECT u.client_id, u.gb, t.price_per_gb
  FROM usage u
  LEFT JOIN LATERAL (
    SELECT ct.tariff_code
    FROM client_tariffs ct
    WHERE ct.client_id = u.client_id
      AND ct.valid_from <= u.day
      AND (ct.valid_to IS NULL OR u.day < ct.valid_to)
    ORDER BY ct.valid_from DESC, ct.loaded_at DESC
    LIMIT 1
  ) v ON true
  LEFT JOIN tariffs t ON t.code = v.tariff_code AND NOT t.archived
  WHERE u.day >= '2024-06-01' AND u.day < '2024-07-01'
)
SELECT c.name AS client,
       coalesce(sum(p.gb) FILTER (WHERE p.price_per_gb IS NOT NULL), 0) AS billed_gb,
       coalesce(sum(p.gb) FILTER (WHERE p.price_per_gb IS NULL), 0)     AS unbilled_gb,
       round(coalesce(sum(p.gb * p.price_per_gb), 0), 2)                AS amount
FROM clients c
LEFT JOIN priced p ON p.client_id = c.id
GROUP BY c.id, c.name
ORDER BY c.name;
