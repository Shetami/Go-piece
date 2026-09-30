SELECT CASE WHEN GROUPING(m.month) = 1 THEN 'Итого'
            ELSE to_char(m.month, 'YYYY-MM') END AS month,
       count(s.id) FILTER (WHERE s.plan = 'basic') AS basic,
       count(s.id) FILTER (WHERE s.plan = 'pro')   AS pro,
       count(s.id) FILTER (WHERE s.plan = 'team')  AS team,
       count(s.id)                                 AS total
FROM generate_series(timestamp '2024-01-01', timestamp '2024-04-01', interval '1 month') AS m(month)
LEFT JOIN subscriptions s
       ON date_trunc('month', s.created_at) = m.month
      AND NOT s.is_trial
GROUP BY ROLLUP (m.month)
ORDER BY GROUPING(m.month), m.month;
