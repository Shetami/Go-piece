-- Сверка выписки банка с платежами системы.
-- Колонки: status, bank_id, payment_id, bank_amount, system_amount.
-- Порядок: status, bank_id (NULL в конце), payment_id.
SELECT CASE
         WHEN b.id IS NULL THEN 'missing_in_bank'
         WHEN p.id IS NULL THEN 'missing_in_system'
         WHEN b.amount <> p.amount THEN 'amount_mismatch'
         ELSE 'matched'
       END AS status,
       b.id AS bank_id, p.id AS payment_id, b.amount AS bank_amount, p.amount AS system_amount
FROM bank_statement b
FULL JOIN payments p ON p.reference = b.reference
ORDER BY status, bank_id NULLS LAST, payment_id;
