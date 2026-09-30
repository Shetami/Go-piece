-- Расхождения между журналом и провайдером.
-- Колонки: external_id, ledger_amount, provider_amount, issue.
-- issue: 'amount_mismatch', 'manual_entry', 'missing_at_provider', 'missing_in_ledger'.
-- Порядок: по issue, external_id, ledger_amount.
SELECT l.external_id, l.amount AS ledger_amount, p.amount AS provider_amount,
       CASE WHEN p.external_id IS NULL THEN 'missing_at_provider'
            WHEN l.external_id IS NULL THEN 'missing_in_ledger'
            ELSE 'amount_mismatch' END AS issue
FROM ledger l
FULL JOIN provider_events p ON p.external_id = l.external_id AND p.status = 'succeeded'
WHERE l.amount <> p.amount OR l.id IS NULL OR p.event_id IS NULL
ORDER BY issue, external_id, ledger_amount;
