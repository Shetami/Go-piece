CREATE TABLE users (
  id     int PRIMARY KEY,
  name   text NOT NULL,
  is_bot boolean NOT NULL DEFAULT false
);

-- В эксперимент попали не все пользователи
CREATE TABLE experiment (
  user_id int PRIMARY KEY REFERENCES users(id),
  variant text NOT NULL -- A или B
);

CREATE TABLE events (
  user_id int NOT NULL REFERENCES users(id),
  ts      timestamp NOT NULL,
  event   text NOT NULL -- view, cart, purchase
);

INSERT INTO users VALUES
  (1, 'Аня',  false),
  (2, 'Боря', false),
  (3, 'Вера', false),
  (4, 'Гоша', false),
  (5, 'Дина', false),
  (6, 'crawler', true);

INSERT INTO experiment VALUES
  (1, 'A'), (2, 'A'),
  (3, 'B'), (4, 'B'), (6, 'B');

INSERT INTO events VALUES
  (1, '2024-05-01 10:00', 'view'),
  (1, '2024-05-01 10:10', 'cart'),
  (1, '2024-05-01 10:40', 'purchase'),
  (1, '2024-05-01 12:00', 'view'),
  (1, '2024-05-01 12:05', 'view'),

  (2, '2024-05-01 09:00', 'cart'),
  (2, '2024-05-01 09:05', 'view'),
  (2, '2024-05-01 09:20', 'purchase'),
  (2, '2024-05-01 09:55', 'cart'),
  (2, '2024-05-01 10:00', 'purchase'),

  (3, '2024-05-01 20:00', 'view'),
  (3, '2024-05-01 20:29', 'cart'),
  (3, '2024-05-01 20:58', 'view'),
  (3, '2024-05-01 21:27', 'purchase'),
  (3, '2024-05-02 20:00', 'view'),
  (3, '2024-05-02 20:50', 'cart'),

  (4, '2024-05-03 10:00', 'view'),
  (4, '2024-05-03 10:15', 'purchase'),
  (4, '2024-05-03 10:20', 'cart'),
  (4, '2024-05-03 10:25', 'purchase'),

  (5, '2024-05-03 11:00', 'view'),
  (5, '2024-05-03 11:01', 'cart'),
  (5, '2024-05-03 11:02', 'purchase'),

  (6, '2024-05-03 12:00', 'view'),
  (6, '2024-05-03 12:00:01', 'cart'),
  (6, '2024-05-03 12:00:02', 'purchase');
