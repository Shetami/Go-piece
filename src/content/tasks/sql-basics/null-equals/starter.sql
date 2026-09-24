SELECT id, name, email
FROM users
WHERE deleted_at = NULL
ORDER BY id;
