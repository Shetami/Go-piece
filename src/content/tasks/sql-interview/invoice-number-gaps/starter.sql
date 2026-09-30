-- Пропуски в нумерации счетов по филиалам и годам.
-- Колонки: branch, year, gap_from, gap_to, missing. Порядок: branch, year, gap_from.
WITH with_next AS (
  SELECT branch_id, issued_on, number,
         lead(number) OVER (PARTITION BY branch_id ORDER BY number) AS next
  FROM invoices
  WHERE status = 'issued'
)
SELECT b.name AS branch, extract(year FROM w.issued_on)::int AS year,
       w.number + 1 AS gap_from, w.next - 1 AS gap_to, w.next - w.number - 1 AS missing
FROM with_next w
JOIN branches b ON b.id = w.branch_id
WHERE w.next - w.number > 1
ORDER BY branch, year, gap_from;
