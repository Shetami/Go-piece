WITH rub AS (
  SELECT coalesce(p.category, 'Без категории') AS category,
         s.amount * coalesce(r.rate_to_rub, 1) AS amount_rub
  FROM sales s
  JOIN products p ON p.id = s.product_id
  LEFT JOIN rates r ON r.currency = upper(s.currency)
  WHERE s.status = 'completed'
)
SELECT category,
       round(sum(amount_rub), 2) AS revenue_rub,
       round(100 * sum(amount_rub) / sum(sum(amount_rub)) OVER (), 1) AS share_pct
FROM rub
GROUP BY category
ORDER BY revenue_rub DESC, category;
