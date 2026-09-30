-- Итоговая таблица олимпиады по правилам ICPC.
-- Колонки: place, team, solved, penalty. Порядок: place, team.
SELECT row_number() OVER (ORDER BY count(DISTINCT s.problem_id) FILTER (WHERE s.verdict = 'OK') DESC) AS place,
       t.name AS team,
       count(DISTINCT s.problem_id) FILTER (WHERE s.verdict = 'OK') AS solved,
       20 * count(*) FILTER (WHERE s.verdict <> 'OK') AS penalty
FROM teams t
JOIN submissions s ON s.team_id = t.id
GROUP BY t.name
ORDER BY place, team;
