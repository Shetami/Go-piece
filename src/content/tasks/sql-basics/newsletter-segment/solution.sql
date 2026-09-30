SELECT DISTINCT ON (lower(trim(c.email)))
       lower(trim(c.email)) AS email,
       c.id                 AS customer_id
FROM customers c
WHERE nullif(trim(c.email), '') IS NOT NULL       -- ни NULL, ни пустых строк, ни пробелов
  AND c.is_blocked IS NOT TRUE                     -- NULL значит «не заблокирован»
  AND c.country IS DISTINCT FROM 'BY'              -- неизвестная страна — не BY
  AND NOT EXISTS (                                 -- NOT IN сломался бы на NULL в bounces
        SELECT 1 FROM bounces b
        WHERE lower(trim(b.email)) = lower(trim(c.email)))
ORDER BY lower(trim(c.email)), c.id;               -- из дублей остаётся младший id
