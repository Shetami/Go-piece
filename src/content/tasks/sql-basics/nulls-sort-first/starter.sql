-- Три самых недавно заходивших пользователя
SELECT name, last_login
FROM users
ORDER BY last_login DESC, name
LIMIT 3;
