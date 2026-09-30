CREATE TABLE order_items (
  id           int PRIMARY KEY,
  category     text,          -- из карточки товара, заполнена как попало
  qty          int NOT NULL,  -- 0: позицию отменили до отгрузки
  refunded_qty int            -- NULL: возвратов не было
);

INSERT INTO order_items VALUES
  (1,  'books',        3, 1),
  (2,  'Books ',       2, NULL),
  (3,  'toys',         4, 0),
  (4,  NULL,           1, 1),
  (5,  '',             2, 0),
  (6,  'toys',         6, 3),
  (7,  'garden',       0, 0),
  (8,  'electronics',  5, NULL),
  (9,  'Electronics',  3, 1),
  (10, '  ',           1, NULL),
  (11, 'Kitchen',     10, 2);
