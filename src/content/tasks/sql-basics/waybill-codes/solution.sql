WITH valid AS (
  SELECT upper(trim(code)) AS code
  FROM waybills
  WHERE upper(trim(code)) ~ '^[A-Z]{3}-[0-9]{4}-[0-9]{1,6}$'   -- сначала проверить формат, потом приводить типы
)
SELECT split_part(code, '-', 1)                AS city,
       split_part(code, '-', 2)::int           AS year,
       count(*)                                AS waybills,
       max(split_part(code, '-', 3)::int)      AS last_number   -- число, а не текст: '99' > '100' как строки
FROM valid
GROUP BY 1, 2
ORDER BY city, year;
