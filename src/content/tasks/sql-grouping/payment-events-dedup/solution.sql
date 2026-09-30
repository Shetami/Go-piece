WITH ev AS (
  SELECT DISTINCT ON (event_id) *
  FROM payment_events
  ORDER BY event_id, received_at
)
SELECT received_at::date AS day,
       count(DISTINCT payment_id) FILTER (WHERE status = 'succeeded')  AS payments,
       coalesce(sum(amount) FILTER (WHERE status = 'succeeded'), 0)   AS gross,
       coalesce(sum(amount) FILTER (WHERE status = 'refunded'), 0)    AS refunds,
       coalesce(sum(amount) FILTER (WHERE status = 'succeeded'), 0)
         - coalesce(sum(amount) FILTER (WHERE status = 'refunded'), 0) AS net
FROM ev
WHERE status IN ('succeeded', 'refunded')
GROUP BY received_at::date
ORDER BY day;
