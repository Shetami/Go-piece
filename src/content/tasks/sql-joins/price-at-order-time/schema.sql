-- История цен. Интервал полуоткрытый: [valid_from, valid_to).
-- valid_to пуст у действующей цены.
-- ETL однажды загрузил часть истории дважды — полные дубли строк.
CREATE TABLE price_history (
  product_id int       NOT NULL,
  price      numeric   NOT NULL,
  valid_from timestamp NOT NULL,
  valid_to   timestamp
);

CREATE TABLE orders (
  id         int PRIMARY KEY,
  created_at timestamp NOT NULL
);

-- Цена в позиции не хранится: её берут из истории на момент заказа
CREATE TABLE order_items (
  order_id   int NOT NULL REFERENCES orders (id),
  product_id int NOT NULL,
  qty        int NOT NULL
);

INSERT INTO price_history VALUES
  (1, 100, '2024-01-01 00:00', '2024-02-01 00:00'),
  (1, 120, '2024-02-01 00:00', '2024-03-01 00:00'),
  (1, 90,  '2024-03-01 00:00', NULL),
  (2, 50,  '2024-01-15 00:00', '2024-02-10 12:00'),
  (2, 50,  '2024-01-15 00:00', '2024-02-10 12:00'),
  (2, 55,  '2024-02-10 12:00', NULL),
  (3, 700, '2024-02-01 00:00', NULL);

INSERT INTO orders VALUES
  (10, '2024-01-20 10:00'),
  (11, '2024-02-01 00:00'),
  (12, '2024-02-10 11:59:59'),
  (13, '2024-02-10 12:00'),
  (14, '2024-01-10 09:00'),
  (15, '2024-03-05 15:00');

INSERT INTO order_items VALUES
  (10, 1, 2),
  (10, 2, 1),
  (11, 1, 1),
  (11, 3, 1),
  (12, 2, 4),
  (13, 2, 4),
  (14, 1, 1),
  (14, 2, 1),
  (14, 3, 1),
  (15, 1, 3),
  (15, 3, 1);
