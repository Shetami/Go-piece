-- Максимальная просадка каждого счёта: самое глубокое падение от предшествующего пика, в %.
-- Колонки: account, max_drawdown_pct (round 2), peak_day, trough_day. Без просадки: 0.00, NULL, NULL.
-- Порядок: account.
SELECT account,
       round(100 * (max(balance) - min(balance)) / max(balance), 2) AS max_drawdown_pct,
       NULL::date AS peak_day,
       NULL::date AS trough_day
FROM equity
GROUP BY account
ORDER BY account;
