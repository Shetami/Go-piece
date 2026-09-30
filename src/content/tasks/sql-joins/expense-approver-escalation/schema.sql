CREATE TABLE employees (
  id         int PRIMARY KEY,
  name       text NOT NULL,
  manager_id int REFERENCES employees (id)
);

-- Отпуска и больничные, обе даты включительно.
-- Отпуск и больничный одного человека могут пересекаться.
CREATE TABLE absences (
  employee_id int  NOT NULL REFERENCES employees (id),
  from_date   date NOT NULL,
  to_date     date NOT NULL,
  kind        text NOT NULL
);

CREATE TABLE expense_requests (
  id          int PRIMARY KEY,
  employee_id int     NOT NULL REFERENCES employees (id),
  created_on  date    NOT NULL,
  amount      numeric NOT NULL
);

INSERT INTO employees VALUES
  (1, 'Ирина',  NULL),
  (2, 'Кирилл', 1),
  (3, 'Лена',   2),
  (4, 'Марат',  2),
  (5, 'Нина',   1),
  (6, 'Олег',   5);

INSERT INTO absences VALUES
  (2, '2024-10-07', '2024-10-11', 'vacation'),
  (2, '2024-10-10', '2024-10-14', 'sick'),
  (1, '2024-10-14', '2024-10-14', 'vacation'),
  (5, '2024-10-01', '2024-10-31', 'vacation'),
  (1, '2024-10-20', '2024-10-25', 'vacation'),
  (3, '2024-10-01', '2024-10-31', 'vacation');

INSERT INTO expense_requests VALUES
  (100, 3, '2024-10-04', 1200),
  (101, 3, '2024-10-07', 800),
  (102, 4, '2024-10-11', 450),
  (103, 4, '2024-10-14', 3000),
  (104, 4, '2024-10-15', 150),
  (105, 6, '2024-10-18', 900),
  (106, 6, '2024-10-22', 700),
  (107, 1, '2024-10-02', 5000),
  (108, 5, '2024-10-03', 200);
