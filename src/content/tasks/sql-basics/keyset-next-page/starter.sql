-- Следующие 6 активных товаров после курсора (rating = 4.5, id = 12).
-- Колонки: id, name, rating.
-- Порядок: rating по убыванию, товары без оценки — в конце, при равенстве — id по возрастанию.
SELECT id, name, rating
FROM products
WHERE is_active
  AND (rating, id) < (4.5, 12)
ORDER BY rating DESC, id
LIMIT 6;
