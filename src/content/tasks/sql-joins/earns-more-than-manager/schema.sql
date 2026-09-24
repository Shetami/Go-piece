CREATE TABLE employees (
  id         int PRIMARY KEY,
  name       text NOT NULL,
  salary     int NOT NULL,
  manager_id int REFERENCES employees (id)
);

INSERT INTO employees VALUES
  (1, 'Анна',   300, NULL),
  (2, 'Борис',  200, 1),
  (3, 'Вика',   250, 2),
  (4, 'Глеб',   200, 2),
  (5, 'Дина',   180, 1),
  (6, 'Егор',   190, 5),
  (7, 'Жанна',  150, 5);
