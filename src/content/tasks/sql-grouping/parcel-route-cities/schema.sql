CREATE TABLE parcels (
  id       int PRIMARY KEY,
  tracking text NOT NULL
);

-- Сканы приходят от сортировочных центров с задержкой: id не отражает порядок.
-- Время сканов у одной посылки не повторяется. city IS NULL — скан без адреса.
CREATE TABLE scans (
  id         int PRIMARY KEY,
  parcel_id  int NOT NULL REFERENCES parcels (id),
  scanned_at timestamp NOT NULL,
  city       text
);

INSERT INTO parcels VALUES (1, 'TRK-001'), (2, 'TRK-002'), (3, 'TRK-003'), (4, 'TRK-004');

INSERT INTO scans VALUES
  (1,  1, '2024-07-01 18:00', 'Тверь'),
  (2,  1, '2024-07-02 15:00', 'Санкт-Петербург'),
  (3,  1, '2024-07-01 12:00', 'Москва'),
  (4,  1, '2024-07-02 09:00', NULL),
  (5,  1, '2024-07-01 10:00', 'Москва'),
  (6,  2, '2024-07-03 08:00', 'Казань'),
  (7,  2, '2024-07-03 12:00', 'Москва'),
  (8,  2, '2024-07-03 20:00', 'Казань'),
  (9,  3, '2024-07-04 10:00', NULL),
  (10, 3, '2024-07-04 11:00', NULL);
