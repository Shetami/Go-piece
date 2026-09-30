WITH suspects AS (
  SELECT c.*
  FROM charges c
  WHERE c.status = 'ok'
    AND EXISTS (
      SELECT 1
      FROM charges p
      WHERE p.merchant_id = c.merchant_id
        AND p.card        = c.card
        AND p.amount      = c.amount
        AND p.status      = 'ok'
        AND p.charged_at >= c.charged_at - interval '5 minutes'
        AND (p.charged_at, p.id) < (c.charged_at, c.id)
    )
)
SELECT m.name,
       count(s.id)                  AS suspects,
       coalesce(sum(s.amount), 0)   AS to_refund
FROM merchants m
LEFT JOIN suspects s ON s.merchant_id = m.id
GROUP BY m.id, m.name
ORDER BY to_refund DESC, m.name;
