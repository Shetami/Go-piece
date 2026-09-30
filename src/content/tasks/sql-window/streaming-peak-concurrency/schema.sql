CREATE TABLE streams (
  id       int PRIMARY KEY,
  title    text      NOT NULL,
  ended_at timestamp NOT NULL
);

-- ended_at IS NULL — зритель досмотрел до конца трансляции (сессия не закрылась).
CREATE TABLE views (
  id         int PRIMARY KEY,
  stream_id  int       NOT NULL REFERENCES streams (id),
  user_id    int       NOT NULL,
  started_at timestamp NOT NULL,
  ended_at   timestamp
);

INSERT INTO streams VALUES
  (1, 'Финал',  '2024-06-01 22:00'),
  (2, 'Разбор', '2024-06-01 19:00'),
  (3, 'Тест',   '2024-06-01 12:00');

INSERT INTO views VALUES
  (1, 1, 101, '2024-06-01 20:00', '2024-06-01 20:30'),
  (2, 1, 102, '2024-06-01 20:10', NULL),
  (3, 1, 103, '2024-06-01 20:30', '2024-06-01 21:00'),  -- зашёл в ту минуту, когда 101 вышел
  (4, 1, 104, '2024-06-01 20:20', '2024-06-01 20:40'),
  (5, 1, 105, '2024-06-01 21:30', '2024-06-01 21:45'),
  (6, 1, 106, '2024-06-01 21:30', NULL),
  (7, 2, 201, '2024-06-01 18:00', '2024-06-01 18:10'),
  (8, 2, 202, '2024-06-01 18:10', '2024-06-01 18:20'),
  (9, 2, 203, '2024-06-01 18:15', NULL);
