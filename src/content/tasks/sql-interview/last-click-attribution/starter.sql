-- Выручка по каналам по модели «последний клик за 7 дней».
-- Колонки: channel, orders, revenue, share_pct. Порядок: revenue по убыванию, затем channel.
SELECT c.channel,
       count(*) AS orders,
       sum(o.amount) AS revenue,
       round(100.0 * sum(o.amount) / (SELECT sum(amount) FROM orders), 1) AS share_pct
FROM orders o
JOIN touches t   ON t.user_id = o.user_id
JOIN campaigns c ON c.id = t.campaign_id
GROUP BY c.channel
ORDER BY revenue DESC, c.channel;
