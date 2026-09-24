SELECT customer,
       count(*)                                          AS orders,
       count(*) FILTER (WHERE status = 'paid')           AS paid,
       count(*) FILTER (WHERE status = 'cancelled')      AS cancelled,
       coalesce(sum(amount) FILTER (WHERE status = 'paid'), 0) AS paid_amount
FROM orders
GROUP BY customer
ORDER BY customer;
