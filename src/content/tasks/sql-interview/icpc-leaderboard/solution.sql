WITH valid AS (
  SELECT s.*, c.starts_at
  FROM submissions s
  CROSS JOIN contest c
  WHERE s.submitted_at >= c.starts_at
    AND s.submitted_at <  c.ends_at
),
first_ok AS (
  SELECT team_id, problem_id, min(submitted_at) AS ok_at, min(starts_at) AS starts_at
  FROM valid
  WHERE verdict = 'OK'
  GROUP BY team_id, problem_id
),
solved AS (
  SELECT f.team_id, f.problem_id,
         floor(extract(epoch FROM f.ok_at - f.starts_at) / 60)::int
           + 20 * (SELECT count(*) FROM valid v
                   WHERE v.team_id = f.team_id
                     AND v.problem_id = f.problem_id
                     AND v.submitted_at < f.ok_at
                     AND v.verdict NOT IN ('OK', 'CE')) AS penalty
  FROM first_ok f
),
totals AS (
  SELECT t.id, t.name,
         count(s.problem_id)        AS solved,
         coalesce(sum(s.penalty), 0) AS penalty
  FROM teams t
  LEFT JOIN solved s ON s.team_id = t.id
  WHERE t.is_official
  GROUP BY t.id, t.name
)
SELECT rank() OVER (ORDER BY solved DESC, penalty) AS place,
       name AS team, solved, penalty
FROM totals
ORDER BY place, team;
