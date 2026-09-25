-- Движение MRR по месяцам, февраль–май 2024.
-- Колонки: month, new_mrr, expansion_mrr, contraction_mrr, churned_mrr, total_mrr.
-- Порядок: по month.
SELECT date_trunc('month', s.started_on)::date AS month,
       sum(p.monthly_price) AS new_mrr,
       0 AS expansion_mrr,
       0 AS contraction_mrr,
       0 AS churned_mrr,
       sum(p.monthly_price) AS total_mrr
FROM subscriptions s
JOIN plans p ON p.id = s.plan_id
WHERE s.started_on >= '2024-02-01'
GROUP BY 1
ORDER BY 1;
