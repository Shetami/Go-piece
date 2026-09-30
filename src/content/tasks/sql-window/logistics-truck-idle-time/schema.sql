CREATE TABLE trucks (
  plate text PRIMARY KEY
);

CREATE TABLE trips (
  id          int PRIMARY KEY,
  truck       text      NOT NULL REFERENCES trucks (plate),
  departed_at timestamp NOT NULL,
  arrived_at  timestamp NOT NULL
);

INSERT INTO trucks VALUES ('T1'), ('T2'), ('T3'), ('T4');

INSERT INTO trips VALUES
  -- T1: id не совпадает с порядком рейсов
  (1, 'T1', '2024-04-02 08:00', '2024-04-02 10:00'),
  (2, 'T1', '2024-04-02 13:00', '2024-04-02 15:00'),
  (3, 'T1', '2024-04-02 10:30', '2024-04-02 12:00'),
  -- T2: второй рейс записан с выездом раньше приезда (ошибка диспетчера)
  (4, 'T2', '2024-04-02 07:00', '2024-04-02 09:00'),
  (5, 'T2', '2024-04-02 08:30', '2024-04-02 11:00'),
  (6, 'T2', '2024-04-02 14:00', '2024-04-02 15:00'),
  -- T3: один рейс
  (7, 'T3', '2024-04-02 09:00', '2024-04-02 10:00');
  -- T4 в этот день не выезжал
