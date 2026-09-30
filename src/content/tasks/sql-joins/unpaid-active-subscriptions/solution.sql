SELECT c.name AS customer, s.id AS subscription_id, s.plan
FROM subscriptions s
JOIN customers c ON c.id = s.customer_id
WHERE s.started_on <= date '2024-06-30'
  AND (s.ended_on IS NULL OR s.ended_on >= date '2024-06-01')
  AND NOT EXISTS (
    SELECT 1
    FROM invoices i
    WHERE i.subscription_id = s.id
      AND i.period_start = date '2024-06-01'
      AND i.paid_at IS NOT NULL
  )
ORDER BY s.id;
