-- Следующая страница каталога после курсора (rating = 4.5, id = 12)
SELECT id, name, rating
FROM products
WHERE is_active
  AND (   rating < 4.5                    -- рейтинг ниже курсора
       OR (rating = 4.5 AND id > 12)      -- та же оценка, но дальше по тай-брейкеру
       OR rating IS NULL)                 -- без оценки — в самом конце выдачи
ORDER BY rating DESC NULLS LAST, id
LIMIT 6;
