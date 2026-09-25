CREATE TABLE students (
  id   int PRIMARY KEY,
  name text NOT NULL
);

CREATE TABLE courses (
  id    int PRIMARY KEY,
  title text NOT NULL
);

CREATE TABLE lessons (
  id          int PRIMARY KEY,
  course_id   int NOT NULL REFERENCES courses(id),
  title       text NOT NULL,
  is_optional boolean NOT NULL DEFAULT false
);

CREATE TABLE enrollments (
  student_id int NOT NULL REFERENCES students(id),
  course_id  int NOT NULL REFERENCES courses(id),
  PRIMARY KEY (student_id, course_id)
);

-- Попыток сдать урок может быть сколько угодно
CREATE TABLE submissions (
  id           int PRIMARY KEY,
  student_id   int NOT NULL REFERENCES students(id),
  lesson_id    int NOT NULL REFERENCES lessons(id),
  score        int NOT NULL,
  submitted_at date NOT NULL
);

INSERT INTO students VALUES
  (1, 'Аня'), (2, 'Боря'), (3, 'Вера'), (4, 'Гоша');

INSERT INTO courses VALUES
  (1, 'Go'), (2, 'SQL');

INSERT INTO lessons VALUES
  (1, 1, 'Синтаксис', false),
  (2, 1, 'Горутины',  false),
  (3, 1, 'Каналы',    false),
  (4, 1, 'Бонус',     true),
  (5, 2, 'SELECT',    false),
  (6, 2, 'JOIN',      false);

INSERT INTO enrollments VALUES
  (1, 1), (2, 1), (3, 1), (4, 1),
  (1, 2), (3, 2), (4, 2);

INSERT INTO submissions VALUES
  (1,  1, 1, 80,  '2024-03-01'),
  (2,  1, 2, 50,  '2024-03-02'),
  (3,  1, 2, 70,  '2024-03-05'),
  (4,  1, 3, 90,  '2024-03-06'),
  (5,  2, 1, 100, '2024-03-01'),
  (6,  2, 1, 100, '2024-03-02'),
  (7,  2, 1, 95,  '2024-03-03'),
  (8,  2, 4, 100, '2024-03-04'),
  (9,  3, 1, 60,  '2024-03-02'),
  (10, 3, 2, 65,  '2024-03-03'),
  (11, 3, 3, 59,  '2024-03-04'),
  (12, 3, 3, 88,  '2024-03-07'),
  (13, 3, 3, 95,  '2024-03-09'),
  (14, 4, 1, 70,  '2024-03-01'),
  (15, 4, 2, 70,  '2024-03-02'),
  (16, 4, 3, 70,  '2024-03-06'),
  (17, 1, 5, 90,  '2024-03-10'),
  (18, 1, 6, 40,  '2024-03-11'),
  (19, 3, 5, 70,  '2024-03-12'),
  (20, 3, 6, 75,  '2024-03-12'),
  (21, 4, 6, 100, '2024-03-01'),
  (22, 2, 5, 80,  '2024-03-01'),
  (23, 2, 6, 80,  '2024-03-02');
