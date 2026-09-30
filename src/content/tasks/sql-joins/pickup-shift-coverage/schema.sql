CREATE TABLE points (
  id        int PRIMARY KEY,
  name      text    NOT NULL,
  is_active boolean NOT NULL
);

-- Сколько сотрудников нужно в час [hour_from, hour_to).
-- point_id пуст у правила по умолчанию; правило конкретного пункта его перекрывает.
CREATE TABLE coverage_rules (
  point_id  int REFERENCES points (id),
  hour_from int NOT NULL,
  hour_to   int NOT NULL,
  required  int NOT NULL
);

-- Смены. Выгрузка из графика иногда содержит одну и ту же смену дважды.
CREATE TABLE shifts (
  employee_id int       NOT NULL,
  point_id    int       NOT NULL REFERENCES points (id),
  starts_at   timestamp NOT NULL,
  ends_at     timestamp NOT NULL
);

INSERT INTO points VALUES
  (1, 'Вокзал', true),
  (2, 'Парк',   true),
  (3, 'Склад',  false),
  (4, 'Центр',  true);

INSERT INTO coverage_rules VALUES
  (NULL, 8,  12, 1),
  (NULL, 12, 18, 2),
  (NULL, 18, 20, 1),
  (1,    12, 14, 3);

INSERT INTO shifts VALUES
  (1, 4, '2024-08-05 08:00', '2024-08-05 12:00'),
  (2, 4, '2024-08-05 12:00', '2024-08-05 20:00'),
  (2, 4, '2024-08-05 12:00', '2024-08-05 20:00'),
  (3, 4, '2024-08-05 12:00', '2024-08-05 17:00'),
  (4, 1, '2024-08-05 08:00', '2024-08-05 14:00'),
  (4, 1, '2024-08-05 08:00', '2024-08-05 14:00'),
  (5, 1, '2024-08-05 12:30', '2024-08-05 20:00'),
  (6, 2, '2024-08-05 10:00', '2024-08-05 19:30'),
  (7, 2, '2024-08-06 08:00', '2024-08-06 20:00');
