CREATE TABLE employees (
  id    int PRIMARY KEY,
  name  text NOT NULL,
  email text
);

INSERT INTO employees VALUES
  (1,  'Анна',   'anna@corp.ru'),
  (2,  'Борис',  'BORIS@Corp.ru'),
  (3,  'Вера',   ' vera@corp.ru '),
  (4,  'Гоша',   'gosha@gmail.com'),
  (5,  'Дина',   'dina@Gmail.com'),
  (6,  'Егор',   'egor'),
  (7,  'Жанна',  NULL),
  (8,  'Зоя',    ''),
  (9,  'Илья',   'ilya@mail.ru'),
  (10, 'Кира',   'kira@mail.ru'),
  (11, 'Лев',    'lev@yandex.ru');
