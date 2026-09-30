WITH digits AS (
  SELECT id, regexp_replace(phone, '[^0-9]', '', 'g') AS d   -- NULL остаётся NULL
  FROM customers
),
norm AS (
  SELECT id,
         CASE WHEN length(d) = 11 AND left(d, 1) IN ('7', '8') THEN '+7' || right(d, 10)
              WHEN length(d) = 10 THEN '+7' || d
         END AS phone                                           -- всё остальное — NULL
  FROM digits
)
SELECT phone,
       count(*)                                  AS customers,
       string_agg(id::text, ', ' ORDER BY id)    AS ids
FROM norm
WHERE phone IS NOT NULL
GROUP BY phone
HAVING count(*) > 1
ORDER BY customers DESC, phone;
