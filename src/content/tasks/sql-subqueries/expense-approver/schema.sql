CREATE TABLE employees (
  id             int PRIMARY KEY,
  name           text NOT NULL,
  manager_id     int REFERENCES employees (id),
  approval_limit numeric(12, 2)    -- NULL: права согласовывать расходы нет
);

CREATE TABLE expenses (
  id          int PRIMARY KEY,
  employee_id int NOT NULL REFERENCES employees (id),
  amount      numeric(12, 2) NOT NULL
);

INSERT INTO employees VALUES
  (1, 'Анна',  NULL, 500000.00),
  (2, 'Борис', 1,    100000.00),
  (3, 'Вика',  2,    NULL),
  (4, 'Глеб',  3,     20000.00),
  (5, 'Дина',  4,    NULL),
  (6, 'Егор',  2,      5000.00),
  (7, 'Зоя',   6,    NULL);

INSERT INTO expenses VALUES
  (1, 5,  15000.00),
  (2, 5,  20000.00),
  (3, 5,  50000.00),
  (4, 4,  10000.00),
  (5, 7,   7000.00),
  (6, 2, 300000.00),
  (7, 1,   1000.00),
  (8, 5, 900000.00);
