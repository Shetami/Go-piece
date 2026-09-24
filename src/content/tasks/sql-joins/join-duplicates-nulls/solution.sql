SELECT
  (SELECT count(*) FROM a JOIN b ON a.x = b.x)      AS inner_join,
  (SELECT count(*) FROM a LEFT JOIN b ON a.x = b.x) AS left_join,
  (SELECT count(*) FROM a FULL JOIN b ON a.x = b.x) AS full_join;
