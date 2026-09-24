SELECT m.month::date AS month, coalesce(sum(s.amount), 0) AS revenue
FROM generate_series(date '2024-01-01', date '2024-06-01', interval '1 month') AS m(month)
LEFT JOIN sales s ON date_trunc('month', s.sold_at) = m.month
GROUP BY m.month
ORDER BY m.month;
