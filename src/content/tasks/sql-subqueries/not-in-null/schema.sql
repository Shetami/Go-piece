CREATE TABLE employees (
  id         int PRIMARY KEY,
  name       text NOT NULL,
  manager_id int REFERENCES employees (id)
);

INSERT INTO employees VALUES
  (1, 'Анна',  NULL),
  (2, 'Борис', 1),
  (3, 'Вика',  2),
  (4, 'Глеб',  2),
  (5, 'Дина',  1);
