CREATE TABLE teams (
  id   int PRIMARY KEY,
  name text NOT NULL
);

-- Смена дежурства: [starts_at, ends_at). Смены могут пересекаться и вкладываться.
CREATE TABLE oncall_shifts (
  id           int PRIMARY KEY,
  team_id      int NOT NULL REFERENCES teams(id),
  engineer     text NOT NULL,
  starts_at    timestamp NOT NULL,
  ends_at      timestamp NOT NULL,
  is_cancelled boolean NOT NULL DEFAULT false
);

INSERT INTO teams VALUES (1, 'Платформа'), (2, 'Платежи'), (3, 'Поиск');

INSERT INTO oncall_shifts VALUES
  (1,  1, 'Аня',  '2024-09-01 20:00', '2024-09-03 09:00', false),
  (2,  1, 'Боря', '2024-09-03 09:00', '2024-09-04 09:00', false),
  (3,  1, 'Вера', '2024-09-03 12:00', '2024-09-03 18:00', false),
  (4,  1, 'Гоша', '2024-09-04 12:00', '2024-09-06 09:00', false),
  (5,  1, 'Дина', '2024-09-05 00:00', '2024-09-05 06:00', false),
  (6,  1, 'Аня',  '2024-09-06 09:00', '2024-09-09 09:00', true),
  (7,  1, 'Боря', '2024-09-06 10:00', '2024-09-10 00:00', false),
  (8,  2, 'Ира',  '2024-09-02 00:00', '2024-09-05 00:00', false),
  (9,  2, 'Олег', '2024-09-05 08:00', '2024-09-08 20:00', false),
  (10, 2, 'Ира',  '2024-09-10 00:00', '2024-09-11 00:00', false);
