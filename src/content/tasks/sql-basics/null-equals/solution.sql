SELECT id, name, email
FROM users
WHERE deleted_at IS NULL
ORDER BY id;
