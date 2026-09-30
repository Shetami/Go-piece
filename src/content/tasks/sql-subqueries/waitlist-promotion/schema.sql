CREATE TABLE events (
  id       int PRIMARY KEY,
  title    text NOT NULL,
  capacity int  NOT NULL
);

CREATE TABLE registrations (
  id         int PRIMARY KEY,
  event_id   int  NOT NULL REFERENCES events (id),
  person     text NOT NULL,
  status     text NOT NULL,         -- 'confirmed', 'waitlist', 'cancelled'
  created_at timestamp NOT NULL
);

INSERT INTO events VALUES
  (1, 'Go-митап',     3),
  (2, 'SQL-воркшоп',  2),
  (3, 'Хакатон',      5),
  (4, 'Лекция',       2);

INSERT INTO registrations VALUES
  (1,  1, 'Аня',   'confirmed', '2024-09-01 10:00'),
  (2,  1, 'Боря',  'confirmed', '2024-09-01 10:01'),
  (3,  1, 'Вера',  'cancelled', '2024-09-01 10:02'),
  (4,  1, 'Даша',  'waitlist',  '2024-09-01 10:05'),
  (5,  1, 'Гоша',  'waitlist',  '2024-09-01 10:05'),
  (6,  1, 'Егор',  'waitlist',  '2024-09-01 10:10'),
  (7,  2, 'Жора',  'confirmed', '2024-09-02 09:00'),
  (8,  2, 'Зина',  'confirmed', '2024-09-02 09:01'),
  (9,  2, 'Илья',  'confirmed', '2024-09-02 09:02'),
  (10, 2, 'Кира',  'waitlist',  '2024-09-02 09:03'),
  (11, 3, 'Лев',   'confirmed', '2024-09-03 07:00'),
  (12, 3, 'Мира',  'waitlist',  '2024-09-03 09:00'),
  (13, 3, 'Нина',  'waitlist',  '2024-09-03 08:00'),
  (14, 3, 'Олег',  'cancelled', '2024-09-03 06:00');
