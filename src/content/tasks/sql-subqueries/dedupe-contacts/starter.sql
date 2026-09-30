-- Удалить дубли контактов по email (без учёта регистра и пробелов по краям), оставив самый
-- свежий по updated_at (неизвестная дата — самая старая; ничья — больший id). Контакты без email не трогать.
-- Вернуть: email (нормализованный), kept_id, removed (сколько удалено). Только группы с удалениями. Порядок: email.
WITH removed AS (
  DELETE FROM contacts c
  WHERE EXISTS (
    SELECT 1 FROM contacts d
    WHERE d.email = c.email AND d.updated_at > c.updated_at
  )
  RETURNING c.email
)
SELECT email, (SELECT max(id) FROM contacts x WHERE x.email = r.email) AS kept_id, count(*) AS removed
FROM removed r
GROUP BY email
ORDER BY email;
