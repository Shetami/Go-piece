-- Выручка по месяцам за первое полугодие 2024 (январь–июнь).
-- Месяцы без продаж тоже нужны — с выручкой 0.
-- Колонки: month (первое число месяца, тип date), revenue. Порядок: по month.
SELECT date_trunc('month', sold_at)::date AS month, sum(amount) AS revenue
FROM sales
GROUP BY 1
ORDER BY 1;
