CREATE TABLE employees (
  id     int PRIMARY KEY,
  name   text NOT NULL,
  salary int              -- NULL: оклад не внесён в систему
);

INSERT INTO employees VALUES
  (1,  'Анна',   90000),
  (2,  'Борис',  100000),
  (3,  'Вера',   99999),
  (4,  'Гоша',   150000),
  (5,  'Дина',   199999),
  (6,  'Егор',   200000),
  (7,  'Жанна',  250000),
  (8,  'Зоя',    NULL),
  (9,  'Илья',   NULL),
  (10, 'Кира',   120000);
