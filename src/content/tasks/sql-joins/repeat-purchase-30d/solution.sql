WITH firsts AS (
  SELECT customer_id, min(created_at) AS first_at
  FROM orders
  WHERE status = 'completed'
  GROUP BY customer_id
),
flags AS (
  SELECT f.customer_id, f.first_at,
         EXISTS (
           SELECT 1
           FROM orders o
           WHERE o.customer_id = f.customer_id
             AND o.status = 'completed'
             AND o.created_at >  f.first_at
             AND o.created_at <= f.first_at + interval '30 days'
         ) AS repeated
  FROM firsts f
)
SELECT to_char(first_at, 'YYYY-MM')              AS cohort,
       count(*)                                  AS customers,
       count(*) FILTER (WHERE repeated)          AS repeaters,
       round(100.0 * count(*) FILTER (WHERE repeated) / count(*), 1) AS repeat_pct
FROM flags
GROUP BY 1
ORDER BY 1;
