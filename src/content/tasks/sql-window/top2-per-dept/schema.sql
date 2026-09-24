CREATE TABLE employees (
  id     int PRIMARY KEY,
  name   text NOT NULL,
  dept   text NOT NULL,
  salary int NOT NULL
);

INSERT INTO employees VALUES
  (1, 'Анна',  'Бэкенд',   300),
  (2, 'Борис', 'Бэкенд',   250),
  (3, 'Вика',  'Бэкенд',   250),
  (4, 'Глеб',  'Бэкенд',   200),
  (5, 'Дина',  'Фронтенд', 220),
  (6, 'Егор',  'Фронтенд', 240),
  (7, 'Жанна', 'Фронтенд', 180),
  (8, 'Зоя',   'Дизайн',   210);
