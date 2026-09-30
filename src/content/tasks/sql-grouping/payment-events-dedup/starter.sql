-- Дневная выручка по вебхукам платежей.
-- Колонки: day (date), payments, gross, refunds, net = gross - refunds.
-- Порядок: по day.
SELECT received_at::date AS day,
       count(*) FILTER (WHERE status = 'succeeded') AS payments,
       sum(amount) FILTER (WHERE status = 'succeeded') AS gross,
       sum(amount) FILTER (WHERE status = 'refunded') AS refunds,
       sum(amount) FILTER (WHERE status = 'succeeded')
         - sum(amount) FILTER (WHERE status = 'refunded') AS net
FROM payment_events
GROUP BY 1
ORDER BY 1;
