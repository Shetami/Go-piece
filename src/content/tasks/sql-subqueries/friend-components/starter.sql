-- Компоненты связности графа дружбы (дружба взаимна). Компонента = наименьший id в ней.
-- Колонки: component, size, members (имена через ', ' по алфавиту). Порядок: component.
WITH component AS (
  SELECT u.id AS user_id,
         least(u.id, coalesce((SELECT min(f.b) FROM friendships f WHERE f.a = u.id), u.id)) AS component
  FROM users u
)
SELECT c.component, count(*) AS size, string_agg(u.name, ', ' ORDER BY u.name) AS members
FROM component c
JOIN users u ON u.id = c.user_id
GROUP BY c.component
ORDER BY c.component;
