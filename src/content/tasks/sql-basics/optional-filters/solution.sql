SELECT f.name AS filter, o.id AS order_id
FROM saved_filters f
JOIN orders o
  ON  (nullif(trim(f.city), '') IS NULL                    -- поле не заполнено — фильтра нет
       OR lower(trim(o.city)) = lower(trim(f.city)))       -- NULL в заказе при активном фильтре не проходит
  AND (nullif(trim(f.status), '') IS NULL
       OR lower(trim(o.status)) = lower(trim(f.status)))
  AND (f.min_amount IS NULL OR o.amount >= f.min_amount)
  AND (f.date_from IS NULL OR o.created_at >= f.date_from)
  AND (f.date_to IS NULL OR o.created_at < f.date_to + 1)  -- весь последний день
ORDER BY filter, order_id;
