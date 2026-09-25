CREATE TABLE products (
  id   int PRIMARY KEY,
  name text NOT NULL
);

-- Партии закупок: у каждой своя цена
CREATE TABLE purchases (
  id         int PRIMARY KEY,
  product_id int NOT NULL REFERENCES products(id),
  bought_at  date NOT NULL,
  qty        int NOT NULL,
  unit_cost  int NOT NULL
);

-- Продажи никогда не превышают остаток на складе
CREATE TABLE sales (
  id         int PRIMARY KEY,
  product_id int NOT NULL REFERENCES products(id),
  sold_at    date NOT NULL,
  qty        int NOT NULL,
  unit_price int NOT NULL
);

INSERT INTO products VALUES
  (1, 'Кофе'),
  (2, 'Чай'),
  (3, 'Какао'),
  (4, 'Сахар');

-- Партии Кофе внесены в базу не в том порядке, в каком закупались
INSERT INTO purchases VALUES
  (1, 1, '2024-01-01', 10, 100),
  (2, 1, '2024-01-20', 5,  120),
  (3, 1, '2024-01-10', 10, 90),
  (4, 2, '2024-01-05', 20, 50),
  (5, 2, '2024-01-15', 10, 60),
  (6, 3, '2024-01-03', 8,  200);

INSERT INTO sales VALUES
  (1, 1, '2024-01-05', 4,  150),
  (2, 1, '2024-01-12', 8,  150),
  (3, 1, '2024-01-25', 6,  160),
  (4, 2, '2024-01-20', 20, 80);
