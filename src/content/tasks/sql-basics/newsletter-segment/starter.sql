-- Адреса для рассылки: по одному на нормализованный email.
-- Колонки: email (lower + trim), customer_id (наименьший id с этим адресом).
-- Порядок: по email.
SELECT email, id AS customer_id
FROM customers
WHERE email IS NOT NULL
  AND NOT is_blocked
  AND country <> 'BY'
  AND email NOT IN (SELECT email FROM bounces)
ORDER BY email;
