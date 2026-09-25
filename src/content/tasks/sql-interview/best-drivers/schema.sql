CREATE TABLE cities (
  id   int PRIMARY KEY,
  name text NOT NULL
);

CREATE TABLE drivers (
  id      int PRIMARY KEY,
  name    text NOT NULL,
  city_id int NOT NULL REFERENCES cities(id)
);

-- arrived_at — когда машина подъехала; у отменённых поездок NULL
CREATE TABLE rides (
  id          int PRIMARY KEY,
  driver_id   int NOT NULL REFERENCES drivers(id),
  accepted_at timestamp NOT NULL,
  arrived_at  timestamp,
  status      text NOT NULL -- done, cancelled_by_client, cancelled_by_driver
);

-- Оценку ставят не за каждую поездку
CREATE TABLE ratings (
  ride_id int PRIMARY KEY REFERENCES rides(id),
  stars   int NOT NULL CHECK (stars BETWEEN 1 AND 5)
);

INSERT INTO cities VALUES
  (1, 'Москва'),
  (2, 'Казань'),
  (3, 'Сочи');

INSERT INTO drivers VALUES
  (1, 'Артём', 1),
  (2, 'Борис', 1),
  (3, 'Валя',  1),
  (4, 'Глеб',  2),
  (5, 'Дана',  2),
  (6, 'Ева',   3);

INSERT INTO rides VALUES
  -- Артём: подача 4, 6, 10 минут, одна отмена по вине водителя
  (1,  1, '2024-05-02 08:00', '2024-05-02 08:04', 'done'),
  (2,  1, '2024-05-03 09:00', '2024-05-03 09:06', 'done'),
  (3,  1, '2024-05-04 10:00', '2024-05-04 10:10', 'done'),
  (4,  1, '2024-05-05 11:00', NULL,               'cancelled_by_driver'),
  -- Борис: подача 2, 6, 8 минут
  (5,  2, '2024-05-02 08:00', '2024-05-02 08:02', 'done'),
  (6,  2, '2024-05-03 09:00', '2024-05-03 09:06', 'done'),
  (7,  2, '2024-05-04 10:00', '2024-05-04 10:08', 'done'),
  -- Валя: в мае только две поездки, третья — в апреле
  (8,  3, '2024-05-10 12:00', '2024-05-10 12:01', 'done'),
  (9,  3, '2024-05-11 12:00', '2024-05-11 12:02', 'done'),
  (10, 3, '2024-04-30 23:50', '2024-04-30 23:51', 'done'),
  -- Глеб: подача 8, 9, 12, 15 минут, две отмены водителем и одна клиентом
  (11, 4, '2024-05-02 08:00', '2024-05-02 08:08', 'done'),
  (12, 4, '2024-05-03 08:00', '2024-05-03 08:09', 'done'),
  (13, 4, '2024-05-04 08:00', '2024-05-04 08:12', 'done'),
  (14, 4, '2024-05-05 08:00', '2024-05-05 08:15', 'done'),
  (15, 4, '2024-05-06 08:00', NULL,               'cancelled_by_driver'),
  (16, 4, '2024-05-07 08:00', NULL,               'cancelled_by_driver'),
  (17, 4, '2024-05-08 08:00', NULL,               'cancelled_by_client'),
  -- Дана: подача 7, 11, 14 минут
  (18, 5, '2024-05-02 08:00', '2024-05-02 08:07', 'done'),
  (19, 5, '2024-05-03 08:00', '2024-05-03 08:11', 'done'),
  (20, 5, '2024-05-04 08:00', '2024-05-04 08:14', 'done'),
  -- Ева: только две поездки
  (21, 6, '2024-05-02 08:00', '2024-05-02 08:05', 'done'),
  (22, 6, '2024-05-03 08:00', '2024-05-03 08:05', 'done');

INSERT INTO ratings VALUES
  (1, 5), (2, 4),
  (5, 5), (6, 5), (7, 5),
  (8, 5), (9, 5), (10, 5),
  (11, 3), (13, 4),
  (18, 5),
  (21, 5), (22, 5);
