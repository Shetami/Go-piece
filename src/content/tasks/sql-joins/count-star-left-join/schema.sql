CREATE TABLE departments (
  id   int PRIMARY KEY,
  name text NOT NULL
);

CREATE TABLE employees (
  id      int PRIMARY KEY,
  name    text NOT NULL,
  dept_id int REFERENCES departments (id)
);

INSERT INTO departments VALUES
  (1, 'Бэкенд'),
  (2, 'Фронтенд'),
  (3, 'Дизайн');

INSERT INTO employees VALUES
  (1, 'Аня',  1),
  (2, 'Боря', 1),
  (3, 'Вера', 1),
  (4, 'Гоша', 2);
