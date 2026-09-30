CREATE TABLE products (
  id   int PRIMARY KEY,
  name text NOT NULL
);

-- История базовых цен: цена действует с valid_from включительно
-- до valid_to не включительно; NULL — действует до сих пор
CREATE TABLE price_history (
  product_id int NOT NULL REFERENCES products(id),
  price      numeric(10,2) NOT NULL,
  valid_from timestamp NOT NULL,
  valid_to   timestamp
);

-- Акции: те же правила границ; акции одного товара могут пересекаться
CREATE TABLE promos (
  product_id  int NOT NULL REFERENCES products(id),
  promo_price numeric(10,2) NOT NULL,
  starts_at   timestamp NOT NULL,
  ends_at     timestamp NOT NULL
);

CREATE TABLE orders (
  id         int PRIMARY KEY,
  status     text NOT NULL, -- paid, cancelled
  created_at timestamp NOT NULL
);

CREATE TABLE order_items (
  order_id   int NOT NULL REFERENCES orders(id),
  product_id int NOT NULL REFERENCES products(id),
  qty        int NOT NULL,
  unit_price numeric(10,2) NOT NULL
);

INSERT INTO products VALUES
  (1, 'Кофе'), (2, 'Чай'), (3, 'Сахар'), (4, 'Молоко');

INSERT INTO price_history VALUES
  (1, 500, '2024-06-01', '2024-06-05'),
  (1, 550, '2024-06-05', NULL),
  (2, 200, '2024-06-01', NULL),
  (3,  90, '2024-06-01', '2024-06-15'),
  (3,  95, '2024-06-15', NULL),
  (4, 120, '2024-06-10', NULL);

INSERT INTO promos VALUES
  (2, 150, '2024-06-03 00:00', '2024-06-04 00:00'),
  (2, 170, '2024-06-03 12:00', '2024-06-06 00:00'),
  (3,  80, '2024-06-10 00:00', '2024-06-12 00:00');

INSERT INTO orders VALUES
  (1, 'paid',      '2024-06-04 23:59:59'),
  (2, 'paid',      '2024-06-05 00:00:00'),
  (3, 'paid',      '2024-06-03 13:00:00'),
  (4, 'paid',      '2024-06-09 18:00:00'),
  (5, 'paid',      '2024-06-12 00:00:00'),
  (6, 'cancelled', '2024-06-05 10:00:00'),
  (7, 'paid',      '2024-06-20 09:00:00');

INSERT INTO order_items VALUES
  (1, 1, 2, 500),
  (1, 2, 1, 170),
  (2, 1, 1, 500),
  (2, 2, 3, 200),
  (3, 2, 2, 170),
  (4, 4, 1, 120),
  (4, 3, 5,  90),
  (5, 3, 2,  80),
  (6, 1, 1,   1),
  (7, 3, 1,  95),
  (7, 1, 1, 550);
