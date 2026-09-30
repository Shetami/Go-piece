CREATE TABLE sellers (
  id   int PRIMARY KEY,
  name text NOT NULL
);

CREATE TABLE categories (
  id   int PRIMARY KEY,
  name text NOT NULL
);

-- Комиссия действует с valid_from до следующей записи категории
CREATE TABLE commission_rates (
  category_id int NOT NULL REFERENCES categories(id),
  valid_from  date NOT NULL,
  pct         numeric(4,1) NOT NULL,
  PRIMARY KEY (category_id, valid_from)
);

-- delivered_at NULL — ещё в пути
CREATE TABLE orders (
  id           int PRIMARY KEY,
  seller_id    int NOT NULL REFERENCES sellers(id),
  created_on   date NOT NULL,
  delivered_at timestamp
);

CREATE TABLE order_items (
  id          int PRIMARY KEY,
  order_id    int NOT NULL REFERENCES orders(id),
  category_id int NOT NULL REFERENCES categories(id),
  price       numeric(10,2) NOT NULL,
  qty         int NOT NULL
);

-- Позицию возвращают целиком
CREATE TABLE returns (
  order_item_id int PRIMARY KEY REFERENCES order_items(id),
  returned_on   date NOT NULL
);

-- Позиции, уже выплаченные в прошлых выплатах
CREATE TABLE paid_items (
  order_item_id int PRIMARY KEY REFERENCES order_items(id),
  payout_date   date NOT NULL
);

INSERT INTO sellers VALUES (1, 'ТехноМир'), (2, 'Модный'), (3, 'Новичок');
INSERT INTO categories VALUES (1, 'Электроника'), (2, 'Одежда');

INSERT INTO commission_rates VALUES
  (1, '2024-01-01',  5.0),
  (1, '2024-10-15',  7.0),
  (2, '2024-01-01', 15.0);

INSERT INTO orders VALUES
  (1, 1, '2024-10-10', '2024-10-20 12:00'),
  (2, 1, '2024-10-16', '2024-11-01 23:00'),
  (3, 1, '2024-10-25', '2024-11-05 10:00'),
  (4, 2, '2024-10-01', '2024-10-05 15:00'),
  (5, 2, '2024-10-02', NULL),
  (6, 2, '2024-09-20', '2024-09-25 11:00');

INSERT INTO order_items VALUES
  (1, 1, 1, 1000, 2),
  (2, 2, 1, 3000, 1),
  (3, 3, 1,  500, 1),
  (4, 4, 2, 2000, 1),
  (5, 4, 2, 1500, 2),
  (6, 5, 2,  800, 1),
  (7, 6, 2, 1000, 1),
  (8, 3, 2,  400, 1);

INSERT INTO returns VALUES (5, '2024-10-12'), (8, '2024-11-07');
INSERT INTO paid_items VALUES (7, '2024-10-15');
