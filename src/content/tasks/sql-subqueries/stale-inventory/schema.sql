CREATE TABLE products (
  id   int PRIMARY KEY,
  name text NOT NULL
);

CREATE TABLE stock (
  product_id int  NOT NULL REFERENCES products (id),
  warehouse  text NOT NULL,
  qty        int  NOT NULL,
  PRIMARY KEY (product_id, warehouse)
);

CREATE TABLE sales (
  id         int PRIMARY KEY,
  product_id int  NOT NULL REFERENCES products (id),
  sold_at    timestamp NOT NULL,
  qty        int  NOT NULL,
  kind       text NOT NULL          -- 'sale', 'return'
);

INSERT INTO products VALUES
  (1, 'Зонт'), (2, 'Кепка'), (3, 'Панама'), (4, 'Шарф'), (5, 'Варежки'), (6, 'Плед');

INSERT INTO stock VALUES
  (1, 'A', 5), (1, 'B', 3),
  (2, 'A', 10),
  (3, 'A', 0), (3, 'B', 0),
  (4, 'A', 7),
  (5, 'B', 4),
  (6, 'A', 2);

INSERT INTO sales VALUES
  (1,  1, '2024-04-01 12:00',    1, 'sale'),
  (2,  1, '2024-05-10 15:00',    2, 'sale'),
  (3,  1, '2024-06-20 10:00',    1, 'return'),
  (4,  2, '2024-06-01 00:00',    1, 'sale'),
  (5,  3, '2024-03-03 11:00',    5, 'sale'),
  (6,  5, '2024-02-01 09:00',    1, 'sale'),
  (7,  5, '2024-05-31 23:59:59', 1, 'sale'),
  (8,  6, '2024-06-15 18:00',    1, 'sale');
