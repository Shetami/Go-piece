CREATE TABLE users (
  id   int PRIMARY KEY,
  name text NOT NULL,
  tz   text NOT NULL -- часовой пояс пользователя
);

-- Моменты записаны с часовым поясом, здесь — в UTC
CREATE TABLE workouts (
  user_id  int NOT NULL REFERENCES users(id),
  done_at  timestamptz NOT NULL
);

INSERT INTO users VALUES
  (1, 'Аня',  'Europe/Moscow'),
  (2, 'Боря', 'Asia/Novosibirsk'),
  (3, 'Вера', 'America/New_York'),
  (4, 'Гоша', 'Europe/Moscow'),
  (5, 'Дина', 'Europe/Moscow');

INSERT INTO workouts VALUES
  (1, '2024-03-01 22:30+00'),
  (1, '2024-03-02 10:00+00'),
  (1, '2024-03-03 20:00+00'),
  (1, '2024-03-04 21:30+00'),

  (2, '2024-03-01 18:00+00'),
  (2, '2024-03-02 16:00+00'),
  (2, '2024-03-03 05:00+00'),
  (2, '2024-03-04 03:00+00'),

  (3, '2024-03-01 15:00+00'),
  (3, '2024-03-03 03:00+00'),
  (3, '2024-03-05 12:00+00'),
  (3, '2024-03-07 02:00+00'),

  (5, '2024-03-10 07:00+00'),
  (5, '2024-03-10 08:00+00'),
  (5, '2024-03-10 09:00+00');
