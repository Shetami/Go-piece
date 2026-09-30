-- Телефоны, которые встречаются у нескольких клиентов.
-- Колонки: phone (+7XXXXXXXXXX), customers, ids (id через ', ' по возрастанию).
-- Порядок: customers по убыванию, затем phone.
SELECT regexp_replace(phone, '[^0-9]', '', 'g') AS phone,
       count(*) AS customers,
       string_agg(id::text, ', ') AS ids
FROM customers
GROUP BY 1
HAVING count(*) > 1
ORDER BY customers DESC;
