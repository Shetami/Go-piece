WITH first_pass AS (
  SELECT student_id, lesson_id, min(submitted_at) AS passed_on
  FROM submissions
  WHERE score >= 60
  GROUP BY student_id, lesson_id
),
required AS (
  SELECT course_id, count(*) AS lessons
  FROM lessons
  WHERE NOT is_optional
  GROUP BY course_id
),
progress AS (
  SELECT e.course_id, e.student_id,
         count(*)          AS passed,
         max(fp.passed_on) AS finished_on
  FROM enrollments e
  JOIN lessons l     ON l.course_id = e.course_id AND NOT l.is_optional
  JOIN first_pass fp ON fp.student_id = e.student_id AND fp.lesson_id = l.id
  GROUP BY e.course_id, e.student_id
)
SELECT c.title AS course,
       row_number() OVER (PARTITION BY c.id ORDER BY p.finished_on, s.name) AS place,
       s.name AS student,
       p.finished_on
FROM progress p
JOIN required r ON r.course_id = p.course_id AND r.lessons = p.passed
JOIN courses c  ON c.id = p.course_id
JOIN students s ON s.id = p.student_id
ORDER BY course, place;
