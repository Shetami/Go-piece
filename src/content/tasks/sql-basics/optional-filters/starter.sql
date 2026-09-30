-- Какие заказы находит каждый сохранённый фильтр.
-- Колонки: filter (имя фильтра), order_id.
-- Порядок: filter, затем order_id.
SELECT f.name AS filter, o.id AS order_id
FROM saved_filters f
JOIN orders o
  ON  o.city = coalesce(f.city, o.city)
  AND o.status = coalesce(f.status, o.status)
  AND o.amount >= coalesce(f.min_amount, o.amount)
  AND o.created_at BETWEEN coalesce(f.date_from, o.created_at) AND coalesce(f.date_to, o.created_at)
ORDER BY filter, order_id;
