CREATE TABLE users (
  id      int PRIMARY KEY,
  name    text NOT NULL,
  is_test boolean NOT NULL DEFAULT false
);

CREATE TABLE events (
  user_id int NOT NULL REFERENCES users(id),
  ts      timestamp NOT NULL,
  kind    text NOT NULL
);

INSERT INTO users VALUES
  (1, 'Аня', false), (2, 'Боря', false), (3, 'Вера', false),
  (4, 'Гоша', false), (5, 'Дина', false), (6, 'qa', true);

INSERT INTO events VALUES
  (1, '2024-01-05 10:00', 'open'),
  (2, '2024-01-20 10:00', 'open'),
  (3, '2024-01-31 23:59', 'open'),

  (1, '2024-02-01 00:00', 'open'),
  (1, '2024-02-10 12:00', 'buy'),
  (2, '2024-02-29 23:59', 'open'),
  (4, '2024-02-14 09:00', 'open'),
  (6, '2024-02-15 09:00', 'open'),

  (1, '2024-03-03 10:00', 'open'),
  (3, '2024-03-31 23:59:59', 'open'),
  (4, '2024-03-10 08:00', 'open'),
  (4, '2024-03-11 08:00', 'open'),

  (5, '2024-04-20 18:00', 'open'),
  (6, '2024-04-21 18:00', 'open'),

  (2, '2024-05-02 10:00', 'open'),
  (4, '2024-05-15 10:00', 'buy'),
  (5, '2024-05-31 22:00', 'open');
