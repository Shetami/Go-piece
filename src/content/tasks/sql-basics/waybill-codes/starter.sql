-- Накладные по городам и годам.
-- Колонки: city, year (int), waybills, last_number (int).
-- Порядок: city, затем year.
SELECT split_part(code, '-', 1) AS city,
       split_part(code, '-', 2) AS year,
       count(*) AS waybills,
       max(split_part(code, '-', 3)) AS last_number
FROM waybills
GROUP BY 1, 2
ORDER BY 1, 2;
