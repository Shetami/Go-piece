CREATE TABLE couriers (
  id   int PRIMARY KEY,
  name text NOT NULL
);

CREATE TABLE orders (
  id      int PRIMARY KEY,
  is_test boolean NOT NULL DEFAULT false
);

-- Попытки вручения. Заказ могут передать другому курьеру.
-- Сканер иногда отправляет одну и ту же попытку дважды.
CREATE TABLE attempts (
  order_id     int NOT NULL REFERENCES orders(id),
  courier_id   int NOT NULL REFERENCES couriers(id),
  attempted_at timestamp NOT NULL,
  result       text NOT NULL -- delivered, no_answer, refused
);

INSERT INTO couriers VALUES (1, 'Аня'), (2, 'Боря'), (3, 'Вера'), (4, 'Гоша');

INSERT INTO orders VALUES
  (1, false), (2, false), (3, false), (4, false),
  (5, false), (6, false), (7, true),  (8, false);

INSERT INTO attempts VALUES
  (1, 1, '2024-07-01 10:00', 'delivered'),
  (2, 1, '2024-07-01 11:00', 'no_answer'),
  (2, 1, '2024-07-02 11:00', 'delivered'),
  (3, 2, '2024-07-01 12:00', 'no_answer'),
  (3, 3, '2024-07-02 12:00', 'delivered'),
  (4, 2, '2024-07-01 13:00', 'delivered'),
  (4, 2, '2024-07-01 13:00', 'delivered'),
  (5, 3, '2024-07-03 09:00', 'refused'),
  (6, 3, '2024-07-04 10:00', 'delivered'),
  (7, 4, '2024-07-04 11:00', 'delivered'),
  (8, 3, '2024-07-05 15:00', 'delivered');
