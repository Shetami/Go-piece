-- Инвентаризации: фактический остаток, пересчитанный руками в момент counted_at.
-- Движения, проведённые ровно в counted_at, в пересчёт уже вошли.
CREATE TABLE inventory_counts (
  warehouse_id int       NOT NULL,
  sku          text      NOT NULL,
  counted_at   timestamp NOT NULL,
  qty          int       NOT NULL
);

-- Движения: приход (+) и расход (−)
CREATE TABLE movements (
  warehouse_id int       NOT NULL,
  sku          text      NOT NULL,
  moved_at     timestamp NOT NULL,
  delta        int       NOT NULL
);

-- Запросы аудиторов: какой был остаток на складе в момент at (включительно)
CREATE TABLE stock_checks (
  id           int PRIMARY KEY,
  warehouse_id int       NOT NULL,
  sku          text      NOT NULL,
  at           timestamp NOT NULL
);

INSERT INTO inventory_counts VALUES
  (1, 'A', '2024-04-01 00:00', 100),
  (1, 'A', '2024-04-10 00:00', 90),
  (2, 'A', '2024-04-05 00:00', 7);

INSERT INTO movements VALUES
  (1, 'A', '2024-03-30 12:00', 50),
  (1, 'A', '2024-04-01 00:00', -3),
  (1, 'A', '2024-04-03 10:00', -20),
  (1, 'A', '2024-04-08 15:00', 5),
  (1, 'A', '2024-04-10 00:00', -4),
  (1, 'A', '2024-04-12 09:00', -10),
  (2, 'A', '2024-04-06 11:00', -2),
  (1, 'B', '2024-04-02 08:00', 30),
  (1, 'B', '2024-04-04 08:00', -12),
  (1, 'B', '2024-04-04 08:00', -3);

INSERT INTO stock_checks VALUES
  (1, 1, 'A', '2024-04-03 10:00'),
  (2, 1, 'A', '2024-04-09 00:00'),
  (3, 1, 'A', '2024-04-10 00:00'),
  (4, 1, 'A', '2024-04-15 00:00'),
  (5, 2, 'A', '2024-04-06 12:00'),
  (6, 1, 'B', '2024-04-04 08:00'),
  (7, 1, 'B', '2024-04-01 00:00'),
  (8, 2, 'B', '2024-04-06 12:00'),
  (9, 1, 'A', '2024-03-31 00:00');
