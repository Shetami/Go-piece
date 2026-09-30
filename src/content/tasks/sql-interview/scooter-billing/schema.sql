CREATE TABLE users (
  id            int PRIMARY KEY,
  name          text NOT NULL,
  has_subscription boolean NOT NULL DEFAULT false
);

-- Тариф по времени суток: from_time включительно, to_time не включительно.
-- Интервал может переходить через полночь.
CREATE TABLE tariffs (
  name          text PRIMARY KEY,
  from_time     time NOT NULL,
  to_time       time NOT NULL,
  price_per_min numeric(6,2) NOT NULL
);

CREATE TABLE settings (
  unlock_fee numeric(6,2) NOT NULL
);

-- ended_at NULL — поездка ещё идёт
CREATE TABLE rides (
  id         int PRIMARY KEY,
  user_id    int NOT NULL REFERENCES users(id),
  started_at timestamp NOT NULL,
  ended_at   timestamp
);

INSERT INTO users VALUES
  (1, 'Аня',  true),
  (2, 'Боря', false),
  (3, 'Вера', false),
  (4, 'Гоша', false);

INSERT INTO tariffs VALUES
  ('день', '07:00', '23:00', 9.00),
  ('ночь', '23:00', '07:00', 6.00);

INSERT INTO settings VALUES (50.00);

INSERT INTO rides VALUES
  (1, 2, '2024-07-01 22:55:00', '2024-07-01 23:05:00'),
  (2, 2, '2024-07-01 10:00:00', '2024-07-01 10:03:01'),
  (3, 3, '2024-07-02 06:58:30', '2024-07-02 07:01:30'),
  (4, 3, '2024-07-02 12:00:00', '2024-07-02 12:00:20'),
  (5, 1, '2024-07-03 23:59:00', '2024-07-04 00:02:00'),
  (6, 1, '2024-07-04 08:00:00', NULL),
  (7, 3, '2024-07-02 12:00:30', '2024-07-02 12:01:00'),
  (8, 1, '2024-07-05 09:00:00', '2024-07-05 09:10:00');
