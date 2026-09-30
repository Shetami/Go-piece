CREATE TABLE devices (
  id text PRIMARY KEY
);

-- Счётчик шлёт показания нерегулярно; NULL — пакет пришёл, но значение битое.
CREATE TABLE meter (
  device text      NOT NULL REFERENCES devices (id),
  ts     timestamp NOT NULL,
  value  numeric(8,2)
);

INSERT INTO devices VALUES ('m1'), ('m2');

INSERT INTO meter VALUES
  ('m1', '2024-02-01 00:10', 8),
  ('m1', '2024-02-01 00:50', 7),
  -- 01:00–01:59 тишина
  ('m1', '2024-02-01 02:15', NULL),
  ('m1', '2024-02-01 03:05', 9),
  ('m1', '2024-02-01 03:40', NULL),
  -- 04:00–04:59 тишина
  ('m1', '2024-02-01 05:30', 4),
  ('m2', '2024-02-01 02:30', 12),
  ('m2', '2024-02-01 04:10', 11),
  ('m2', '2024-02-01 06:00', 99);   -- уже за пределами отчёта
