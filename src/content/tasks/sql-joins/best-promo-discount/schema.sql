CREATE TABLE products (
  id          int PRIMARY KEY,
  name        text NOT NULL,
  category_id int  NOT NULL
);

CREATE TABLE orders (
  id         int PRIMARY KEY,
  created_at timestamp NOT NULL
);

CREATE TABLE order_items (
  id         int PRIMARY KEY,
  order_id   int     NOT NULL REFERENCES orders (id),
  product_id int     NOT NULL REFERENCES products (id),
  qty        int     NOT NULL,
  price      numeric NOT NULL
);

-- Акции: действуют на [starts_at, ends_at), ends_at пуст — бессрочно.
-- category_id пуст — акция на весь каталог. Акции могут пересекаться.
CREATE TABLE promotions (
  id           int PRIMARY KEY,
  category_id  int,
  discount_pct numeric   NOT NULL,
  starts_at    timestamp NOT NULL,
  ends_at      timestamp
);

INSERT INTO products VALUES
  (1, 'Палатка',   10),
  (2, 'Спальник',  10),
  (3, 'Фонарь',    20),
  (4, 'Котелок',   30);

INSERT INTO orders VALUES
  (100, '2024-07-01 12:00'),
  (101, '2024-07-05 00:00'),
  (102, '2024-07-10 23:59'),
  (103, '2024-06-30 23:59'),
  (104, '2024-07-20 10:00');

INSERT INTO order_items VALUES
  (1, 100, 1, 1, 10000),
  (2, 100, 3, 2, 1500),
  (3, 101, 2, 1, 4000),
  (4, 101, 4, 1, 1000),
  (5, 102, 1, 1, 10000),
  (6, 102, 3, 1, 1500),
  (7, 103, 1, 1, 10000),
  (8, 104, 4, 3, 1000);

INSERT INTO promotions VALUES
  (1, 10,   10, '2024-07-01 00:00', '2024-07-05 00:00'),
  (2, NULL, 5,  '2024-07-01 00:00', '2024-07-11 00:00'),
  (3, 10,   15, '2024-07-05 00:00', '2024-07-08 00:00'),
  (4, 20,   20, '2024-07-01 00:00', NULL),
  (5, 20,   3,  '2024-07-01 00:00', '2024-08-01 00:00'),
  (6, 40,   50, '2024-07-01 00:00', NULL);
