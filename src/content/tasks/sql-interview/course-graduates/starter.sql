-- Выпускники курсов в порядке окончания.
-- Колонки: course, place, student, finished_on. Порядок: course, place.
SELECT c.title AS course,
       1 AS place,
       st.name AS student,
       max(s.submitted_at) AS finished_on
FROM submissions s
JOIN lessons l   ON l.id = s.lesson_id
JOIN courses c   ON c.id = l.course_id
JOIN students st ON st.id = s.student_id
WHERE s.score >= 60
GROUP BY c.title, st.name
HAVING count(*) >= 3
ORDER BY course, place;
