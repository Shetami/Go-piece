WITH moves AS (
  SELECT to_account AS account_id, amount
  FROM transactions
  WHERE to_account IS NOT NULL AND ts < '2024-04-01'
  UNION ALL
  SELECT from_account, -amount
  FROM transactions
  WHERE from_account IS NOT NULL AND ts < '2024-04-01'
),
balances AS (
  SELECT account_id, sum(amount) AS balance
  FROM moves
  GROUP BY account_id
),
rates AS (
  SELECT DISTINCT ON (currency) currency, rub_rate
  FROM fx_rates
  WHERE rate_date <= date '2024-03-31'
  ORDER BY currency, rate_date DESC
),
per_client AS (
  SELECT c.name,
         count(a.id) AS accounts,
         coalesce(sum(b.balance * CASE WHEN a.currency = 'RUB' THEN 1 ELSE r.rub_rate END), 0) AS balance_rub
  FROM clients c
  LEFT JOIN accounts a ON a.client_id = c.id
  LEFT JOIN balances b ON b.account_id = a.id
  LEFT JOIN rates r    ON r.currency = a.currency
  GROUP BY c.id, c.name
)
SELECT name AS client,
       accounts,
       round(balance_rub, 2) AS balance_rub,
       rank() OVER (ORDER BY balance_rub DESC) AS place
FROM per_client
ORDER BY place, client;
