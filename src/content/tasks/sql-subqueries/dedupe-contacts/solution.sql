WITH ranked AS (
  SELECT id,
         lower(trim(email)) AS email,
         row_number() OVER (PARTITION BY lower(trim(email))
                            ORDER BY updated_at DESC NULLS LAST, id DESC) AS rn
  FROM contacts
  WHERE email IS NOT NULL
),
removed AS (
  DELETE FROM contacts c
  USING ranked r
  WHERE c.id = r.id
    AND r.rn > 1
  RETURNING c.id
)
SELECT r.email,
       max(r.id) FILTER (WHERE r.rn = 1) AS kept_id,
       count(d.id)                       AS removed
FROM ranked r
LEFT JOIN removed d ON d.id = r.id
GROUP BY r.email
HAVING count(d.id) > 0
ORDER BY r.email;
