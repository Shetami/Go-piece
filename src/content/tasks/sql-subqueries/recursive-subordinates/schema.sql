CREATE TABLE employees (
  id         int PRIMARY KEY,
  name       text NOT NULL,
  manager_id int REFERENCES employees (id)
);

INSERT INTO employees VALUES
  (1, 'Анна',   NULL),
  (2, 'Борис',  1),
  (3, 'Вика',   2),
  (4, 'Глеб',   2),
  (5, 'Дина',   3),
  (6, 'Егор',   5),
  (7, 'Жанна',  1),
  (8, 'Зоя',    7),
  (9, 'Илья',   4);
