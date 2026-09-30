CREATE TABLE customers (
  id   int PRIMARY KEY,
  name text NOT NULL
);

CREATE TABLE categories (
  id   int PRIMARY KEY,
  name text NOT NULL
);

-- category_id пуст у товаров, которые ещё не разложили по каталогу
CREATE TABLE products (
  id          int PRIMARY KEY,
  name        text NOT NULL,
  category_id int REFERENCES categories (id)
);

CREATE TABLE orders (
  id          int PRIMARY KEY,
  customer_id int NOT NULL REFERENCES customers (id),
  status      text NOT NULL  -- 'paid', 'delivered', 'cancelled'
);

CREATE TABLE order_items (
  order_id   int NOT NULL REFERENCES orders (id),
  product_id int NOT NULL REFERENCES products (id),
  qty        int NOT NULL
);

INSERT INTO customers VALUES
  (1, 'Аня'),
  (2, 'Боря'),
  (3, 'Вера'),
  (4, 'Аня'),
  (5, 'Гоша'),
  (6, 'Дина');

INSERT INTO categories VALUES
  (10, 'Кофе'),
  (20, 'Чай'),
  (30, 'Посуда');

INSERT INTO products VALUES
  (100, 'Эфиопия, 250 г',   10),
  (101, 'Колумбия, 1 кг',   10),
  (102, 'Сенча',            20),
  (103, 'Кружка',           30),
  (104, 'Кофе без карточки', NULL);

INSERT INTO orders VALUES
  (1, 1, 'delivered'),
  (2, 1, 'paid'),
  (3, 2, 'cancelled'),
  (4, 3, 'delivered'),
  (5, 4, 'paid'),
  (6, 5, 'delivered'),
  (7, 6, 'paid');

INSERT INTO order_items VALUES
  (1, 100, 1),
  (1, 101, 2),
  (2, 100, 1),
  (3, 101, 1),
  (4, 102, 3),
  (4, 103, 1),
  (5, 101, 1),
  (6, 104, 2),
  (7, 103, 1),
  (7, 102, 1);
