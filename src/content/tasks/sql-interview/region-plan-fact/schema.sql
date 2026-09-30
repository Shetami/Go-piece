-- Дерево регионов произвольной глубины
CREATE TABLE regions (
  id        int PRIMARY KEY,
  parent_id int REFERENCES regions(id),
  name      text NOT NULL
);

-- Магазин может быть привязан к узлу любого уровня
CREATE TABLE stores (
  id        int PRIMARY KEY,
  region_id int NOT NULL REFERENCES regions(id),
  name      text NOT NULL
);

-- Возвраты записаны отрицательными суммами
CREATE TABLE sales (
  store_id int NOT NULL REFERENCES stores(id),
  sold_at  timestamp NOT NULL,
  amount   numeric(12,2) NOT NULL
);

-- План ставится на узел дерева и покрывает всё его поддерево
CREATE TABLE plans (
  region_id int NOT NULL REFERENCES regions(id),
  month     date NOT NULL, -- первое число месяца
  amount    numeric(12,2) NOT NULL,
  PRIMARY KEY (region_id, month)
);

INSERT INTO regions VALUES
  (1, NULL, 'Россия'),
  (2, 1,    'Центр'),
  (3, 2,    'Москва'),
  (4, 2,    'Тверь'),
  (5, 1,    'Сибирь'),
  (6, 5,    'Новосибирск'),
  (7, 5,    'Томск');

INSERT INTO stores VALUES
  (1, 3, 'Тверская'),
  (2, 3, 'Арбат'),
  (3, 4, 'Тверь-центр'),
  (4, 6, 'Академгородок'),
  (5, 5, 'Склад-магазин Сибирь');

INSERT INTO plans VALUES
  (1, '2024-04-01', 1000), (1, '2024-05-01', 1000), (1, '2024-06-01', 1200),
  (2, '2024-04-01',  600), (2, '2024-05-01',  600), (2, '2024-06-01',  700),
  (5, '2024-04-01',  300),                          (5, '2024-06-01',  400);

INSERT INTO sales VALUES
  (1, '2024-03-31 23:59', 777),
  (1, '2024-04-05 12:00', 300),
  (1, '2024-04-20 12:00', 200),
  (1, '2024-05-10 12:00', 350),
  (1, '2024-06-15 12:00', 400),
  (1, '2024-06-30 23:30', 100),
  (1, '2024-07-01 00:00', 999),
  (2, '2024-04-10 12:00', 150),
  (2, '2024-05-12 12:00', -50),
  (2, '2024-06-01 00:00', 200),
  (3, '2024-05-03 12:00', 100),
  (3, '2024-06-20 12:00', 120),
  (4, '2024-04-15 12:00', 250),
  (4, '2024-06-10 12:00', 300),
  (5, '2024-04-25 12:00',  80),
  (5, '2024-06-25 12:00',  90);
