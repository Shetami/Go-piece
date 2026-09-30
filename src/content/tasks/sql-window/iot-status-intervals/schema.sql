CREATE TABLE device_status (
  device text      NOT NULL,
  ts     timestamp NOT NULL,
  status text      NOT NULL CHECK (status IN ('online', 'offline'))
);

-- Устройство шлёт статус при изменении и иногда — просто так, повторно.
-- Отчёт строится на 2024-07-01 12:00.
INSERT INTO device_status VALUES
  ('dev1', '2024-07-01 10:00', 'online'),
  ('dev1', '2024-07-01 10:05', 'online'),
  ('dev1', '2024-07-01 10:20', 'offline'),
  ('dev1', '2024-07-01 10:20', 'offline'),
  ('dev1', '2024-07-01 10:30', 'offline'),
  ('dev1', '2024-07-01 10:45', 'online'),
  ('dev1', '2024-07-01 11:30', 'offline'),
  ('dev2', '2024-07-01 09:00', 'offline'),
  ('dev2', '2024-07-01 11:00', 'offline'),
  ('dev3', '2024-07-01 11:50', 'online');
