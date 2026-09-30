-- По каждому мерчанту: сколько успешных списаний похожи на повтор (та же карта, мерчант и сумма,
-- а предыдущее такое успешное списание было не раньше чем за 5 минут) и сколько вернуть.
-- Колонки: name, suspects, to_refund (0, если нечего). Порядок: to_refund DESC, name.
SELECT m.name, count(*) AS suspects, sum(c.amount) AS to_refund
FROM charges c
JOIN charges p ON p.card = c.card AND p.amount = c.amount
              AND p.charged_at < c.charged_at
              AND p.charged_at > c.charged_at - interval '5 minutes'
JOIN merchants m ON m.id = c.merchant_id
GROUP BY m.id, m.name
ORDER BY to_refund DESC, m.name;
