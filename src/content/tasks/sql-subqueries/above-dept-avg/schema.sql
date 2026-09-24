CREATE TABLE employees (
  id     int PRIMARY KEY,
  name   text NOT NULL,
  dept   text NOT NULL,
  salary int NOT NULL
);

INSERT INTO employees VALUES
  (1, 'Анна',  'Бэкенд',   300),
  (2, 'Борис', 'Бэкенд',   210),
  (3, 'Вика',  'Бэкенд',   250),
  (4, 'Глеб',  'Фронтенд', 220),
  (5, 'Дина',  'Фронтенд', 180),
  (6, 'Егор',  'Дизайн',   190),
  (7, 'Жанна', 'Бэкенд',   150),
  (8, 'Зоя',   'Фронтенд', 200),
  (9, 'Ира',   'Дизайн',   130);
