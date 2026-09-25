CREATE TABLE departments (
  id   int PRIMARY KEY,
  name text NOT NULL
);

CREATE TABLE employees (
  id            int PRIMARY KEY,
  name          text NOT NULL,
  manager_id    int REFERENCES employees(id),
  department_id int NOT NULL REFERENCES departments(id),
  hired_at      date NOT NULL,
  fired_at      date -- последний день работы уже в прошлом; NULL — работает
);

-- История окладов: новая строка действует с valid_from до следующей строки
CREATE TABLE salaries (
  employee_id int NOT NULL REFERENCES employees(id),
  amount      int NOT NULL,
  valid_from  date NOT NULL,
  PRIMARY KEY (employee_id, valid_from)
);

INSERT INTO departments VALUES
  (1, 'Управление'),
  (2, 'Разработка'),
  (3, 'Продажи');

INSERT INTO employees VALUES
  (1,  'Ольга',   NULL, 1, '2020-01-01', NULL),
  (2,  'Пётр',    1,    2, '2021-01-01', NULL),
  (3,  'Роман',   1,    3, '2021-01-01', NULL),
  (4,  'Света',   2,    2, '2022-01-01', NULL),
  (5,  'Тимур',   2,    2, '2022-01-01', NULL),
  (6,  'Ульяна',  5,    2, '2022-06-01', NULL),
  (7,  'Фёдор',   5,    2, '2023-01-01', NULL),
  (8,  'Хлоя',    3,    3, '2022-01-01', '2024-03-01'),
  (9,  'Циля',    3,    3, '2022-01-01', NULL),
  (10, 'Чен',     6,    2, '2024-02-01', NULL),
  (11, 'Шура',    3,    3, '2024-06-15', NULL),
  (12, 'Эля',     9,    3, '2024-01-10', NULL);

INSERT INTO salaries VALUES
  (1,  500, '2020-01-01'), (1, 600, '2024-01-01'),
  (2,  300, '2021-01-01'), (2, 350, '2024-07-01'),
  (3,  320, '2021-01-01'),
  (4,  150, '2022-01-01'), (4, 170, '2023-06-01'),
  (5,  250, '2022-01-01'),
  (6,  180, '2022-06-01'), (6, 200, '2024-05-31'),
  (7,  160, '2023-01-01'),
  (8,  140, '2022-01-01'),
  (9,  200, '2022-01-01'),
  (10, 120, '2024-02-01'),
  (11, 130, '2024-06-15'),
  (12, 110, '2024-01-10');
