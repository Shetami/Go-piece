SELECT
  avg(rating),
  sum(rating) / count(*),
  count(rating)
FROM reviews;
