CREATE TABLE employees (
  id     int PRIMARY KEY,
  name   text NOT NULL,
  salary int NOT NULL
);

INSERT INTO employees VALUES
  (1, 'Анна',  300),
  (2, 'Борис', 250),
  (3, 'Вика',  300),
  (4, 'Глеб',  200),
  (5, 'Дина',  250),
  (6, 'Егор',  150);
