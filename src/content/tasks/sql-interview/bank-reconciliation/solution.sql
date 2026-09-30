WITH pay AS (
  SELECT * FROM payments WHERE status = 'captured'
),
by_ref AS (
  SELECT b.id AS bank_id, p.id AS payment_id
  FROM bank_statement b
  JOIN pay p ON upper(trim(b.reference)) = upper(trim(p.reference))
),
rest_bank AS (
  SELECT b.*, row_number() OVER (PARTITION BY amount ORDER BY tx_date, id) AS rn
  FROM bank_statement b
  WHERE b.id NOT IN (SELECT bank_id FROM by_ref)
),
rest_pay AS (
  SELECT p.*, row_number() OVER (PARTITION BY amount ORDER BY paid_at, id) AS rn
  FROM pay p
  WHERE p.id NOT IN (SELECT payment_id FROM by_ref)
),
by_amount AS (
  SELECT b.id AS bank_id, p.id AS payment_id
  FROM rest_bank b
  JOIN rest_pay p ON p.amount = b.amount
                 AND p.rn = b.rn
                 AND abs(b.tx_date - p.paid_at::date) <= 1
),
pairs AS (
  SELECT * FROM by_ref
  UNION ALL
  SELECT * FROM by_amount
)
SELECT CASE
         WHEN b.id IS NULL THEN 'missing_in_bank'
         WHEN p.id IS NULL THEN 'missing_in_system'
         WHEN b.amount <> p.amount THEN 'amount_mismatch'
         ELSE 'matched'
       END AS status,
       b.id AS bank_id,
       p.id AS payment_id,
       b.amount AS bank_amount,
       p.amount AS system_amount
FROM pairs x
FULL JOIN bank_statement b ON b.id = x.bank_id
FULL JOIN pay p            ON p.id = x.payment_id
ORDER BY status, bank_id NULLS LAST, payment_id;
