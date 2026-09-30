WITH numbers AS (
  SELECT DISTINCT branch_id, extract(year FROM issued_on)::int AS year, number
  FROM invoices
),
with_prev AS (
  SELECT n.*,
         lag(number, 1, 0) OVER (PARTITION BY branch_id, year ORDER BY number) AS prev
  FROM numbers n
)
SELECT b.name AS branch, w.year,
       w.prev + 1       AS gap_from,
       w.number - 1     AS gap_to,
       w.number - w.prev - 1 AS missing
FROM with_prev w
JOIN branches b ON b.id = w.branch_id
WHERE w.number - w.prev > 1
ORDER BY branch, w.year, gap_from;
