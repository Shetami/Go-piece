CREATE TABLE managers (
  id   int PRIMARY KEY,
  name text NOT NULL
);

CREATE TABLE sales (
  id         int PRIMARY KEY,
  manager_id int NOT NULL REFERENCES managers(id),
  amount     numeric(12,2) NOT NULL,
  sold_on    date NOT NULL
);

-- Возврат может быть частичным, по одной продаже — несколько возвратов
CREATE TABLE returns (
  id          int PRIMARY KEY,
  sale_id     int NOT NULL REFERENCES sales(id),
  amount      numeric(12,2) NOT NULL,
  returned_on date NOT NULL
);

-- Прогрессивная шкала: процент действует только на часть продаж внутри ступени
CREATE TABLE bonus_tiers (
  from_amount numeric(12,2) NOT NULL,
  to_amount   numeric(12,2),          -- NULL — без верхней границы
  pct         numeric(4,1)  NOT NULL
);

INSERT INTO managers VALUES
  (1, 'Аня'), (2, 'Боря'), (3, 'Вера'), (4, 'Гоша'), (5, 'Дина');

INSERT INTO bonus_tiers VALUES
  (0,      100000, 0.0),
  (100000, 300000, 3.0),
  (300000, NULL,   5.0);

INSERT INTO sales VALUES
  (1,  1, 200000, '2024-04-10'),
  (2,  1, 150000, '2024-05-15'),
  (3,  1,  30000, '2024-06-01'),
  (4,  1,  50000, '2024-03-31'),
  (5,  2,  60000, '2024-04-02'),
  (6,  2,  40000, '2024-06-20'),
  (7,  3, 250000, '2024-05-05'),
  (8,  3,  70000, '2024-06-11'),
  (9,  4, 120000, '2024-06-30'),
  (10, 4, 500000, '2024-07-01');

INSERT INTO returns VALUES
  (1, 3, 10000, '2024-06-03'),
  (2, 3, 20000, '2024-07-02'),
  (3, 8, 20000, '2024-06-15');
