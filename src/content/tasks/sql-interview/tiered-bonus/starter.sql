-- Квартальный бонус менеджеров по прогрессивной шкале, II квартал 2024.
-- Колонки: manager, net_sales, bonus, top_tier_pct. Порядок: bonus по убыванию, затем manager.
SELECT m.name AS manager,
       sum(s.amount) - coalesce(sum(r.amount), 0) AS net_sales,
       round((sum(s.amount) - coalesce(sum(r.amount), 0)) * max(t.pct) / 100, 2) AS bonus,
       max(t.pct) AS top_tier_pct
FROM managers m
JOIN sales s        ON s.manager_id = m.id
LEFT JOIN returns r ON r.sale_id = s.id
JOIN bonus_tiers t  ON s.amount >= t.from_amount
WHERE s.sold_on BETWEEN '2024-04-01' AND '2024-07-01'
GROUP BY m.name
ORDER BY bonus DESC, manager;
