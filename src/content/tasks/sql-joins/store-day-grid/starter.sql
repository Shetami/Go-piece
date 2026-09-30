-- Выручка каждого работавшего магазина за каждый день с 1 по 5 марта 2024 года.
-- Колонки: day, store, checks, revenue. Дни без продаж — с нулями.
-- Порядок: по day, затем по store.
SELECT sold_at::date AS day, s.name AS store, count(*) AS checks, sum(amount) AS revenue
FROM sales x
JOIN stores s ON s.id = x.store_id
WHERE sold_at BETWEEN '2024-03-01' AND '2024-03-06'
GROUP BY 1, 2
ORDER BY day, store;
