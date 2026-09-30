-- Выручка по категориям в рублях и доля от общей выручки.
-- Колонки: category ('Без категории' для NULL), revenue_rub (round 2),
--          share_pct (round 1). Порядок: revenue_rub по убыванию, затем category.
SELECT coalesce(p.category, 'Без категории') AS category,
       round(sum(s.amount * r.rate_to_rub), 2) AS revenue_rub,
       round(100 * sum(s.amount * r.rate_to_rub) / (SELECT sum(amount) FROM sales), 1) AS share_pct
FROM sales s
JOIN products p ON p.id = s.product_id
JOIN rates r ON r.currency = s.currency
GROUP BY p.category
ORDER BY revenue_rub DESC, category;
