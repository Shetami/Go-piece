WITH edges AS (
  SELECT user_a AS src, user_b AS dst FROM connections
  UNION
  SELECT user_b, user_a FROM connections
),
alive AS (
  SELECT e.src, e.dst
  FROM edges e
  JOIN users s ON s.id = e.src AND NOT s.is_deleted
  JOIN users d ON d.id = e.dst AND NOT d.is_deleted
  WHERE e.src <> e.dst
),
candidates AS (
  SELECT f1.src AS user_id, f2.dst AS candidate_id, count(*) AS mutual
  FROM alive f1
  JOIN alive f2 ON f2.src = f1.dst
  WHERE f2.dst <> f1.src
    AND NOT EXISTS (
      SELECT 1 FROM alive x WHERE x.src = f1.src AND x.dst = f2.dst
    )
  GROUP BY f1.src, f2.dst
),
ranked AS (
  SELECT *, row_number() OVER (PARTITION BY user_id ORDER BY mutual DESC, candidate_id) AS rn
  FROM candidates
)
SELECT u.name AS user_name, c.name AS suggestion, r.mutual
FROM ranked r
JOIN users u ON u.id = r.user_id
JOIN users c ON c.id = r.candidate_id
WHERE r.rn <= 2
ORDER BY r.user_id, r.mutual DESC, r.candidate_id;
