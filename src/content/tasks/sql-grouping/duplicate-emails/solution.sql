SELECT lower(email) AS email, count(*) AS users
FROM users
GROUP BY lower(email)
HAVING count(*) > 1
ORDER BY email;
