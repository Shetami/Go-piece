SELECT
  count(*) FILTER (WHERE status = 'cancelled') / count(*) * 100 AS cancelled_pct
FROM orders;
