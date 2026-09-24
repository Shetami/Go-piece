SELECT
  round(100.0 * count(*) FILTER (WHERE status = 'cancelled') / count(*), 1) AS cancelled_pct
FROM orders;
