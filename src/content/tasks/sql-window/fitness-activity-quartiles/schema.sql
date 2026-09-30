CREATE TABLE users (
  id   int PRIMARY KEY,
  name text NOT NULL
);

CREATE TABLE workouts (
  id      int PRIMARY KEY,
  user_id int  NOT NULL REFERENCES users (id),
  day     date NOT NULL,
  minutes int            -- NULL: тренировку добавили вручную, без длительности
);

INSERT INTO users VALUES
  (1, 'Аня'), (2, 'Боря'), (3, 'Вера'), (4, 'Гоша'),
  (5, 'Дина'), (6, 'Егор'), (7, 'Жора');

INSERT INTO workouts VALUES
  (1, 1, '2024-05-01', 100),
  (2, 1, '2024-05-03', 200),
  (3, 2, '2024-05-02', 250),
  (4, 3, '2024-05-02', 150),
  (5, 3, '2024-05-04', 100),
  (6, 4, '2024-05-05', 120),
  (7, 4, '2024-05-06', NULL),
  (8, 6, '2024-05-07', 60),
  (9, 7, '2024-05-08', NULL);
  -- у Дины тренировок нет
