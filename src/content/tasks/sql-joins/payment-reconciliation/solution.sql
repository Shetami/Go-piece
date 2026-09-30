WITH last_state AS (
  SELECT DISTINCT ON (external_id) external_id, status, amount
  FROM provider_events
  ORDER BY external_id, received_at DESC, event_id DESC
),
provider AS (
  SELECT external_id, amount
  FROM last_state
  WHERE status = 'succeeded'
),
book AS (
  SELECT external_id, sum(amount) AS amount
  FROM ledger
  WHERE external_id IS NOT NULL
  GROUP BY external_id
)
SELECT external_id,
       b.amount AS ledger_amount,
       p.amount AS provider_amount,
       CASE
         WHEN b.amount IS NULL THEN 'missing_in_ledger'
         WHEN p.amount IS NULL THEN 'missing_at_provider'
         ELSE 'amount_mismatch'
       END AS issue
FROM book b
FULL JOIN provider p USING (external_id)
WHERE b.amount IS DISTINCT FROM p.amount
UNION ALL
SELECT NULL, amount, NULL, 'manual_entry'
FROM ledger
WHERE external_id IS NULL
ORDER BY issue, external_id, ledger_amount;
