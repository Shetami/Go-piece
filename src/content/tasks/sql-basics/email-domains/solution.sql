SELECT split_part(lower(trim(email)), '@', 2) AS domain,
       count(*)                                AS employees
FROM employees
WHERE split_part(lower(trim(email)), '@', 2) <> ''   -- отсекает и NULL, и '' и адреса без @
GROUP BY 1
ORDER BY employees DESC, domain;
