SELECT
  count(*),
  count(phone),
  count(city),
  count(DISTINCT city)
FROM customers;
