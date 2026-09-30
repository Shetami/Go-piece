CREATE TABLE halls (
  id   int PRIMARY KEY,
  name text NOT NULL
);

-- Нумерация мест в ряду может прерываться проходом
CREATE TABLE seats (
  id        int PRIMARY KEY,
  hall_id   int NOT NULL REFERENCES halls(id),
  row_no    int NOT NULL,
  seat_no   int NOT NULL,
  is_broken boolean NOT NULL DEFAULT false
);

CREATE TABLE screenings (
  id        int PRIMARY KEY,
  hall_id   int NOT NULL REFERENCES halls(id),
  film      text NOT NULL,
  starts_at timestamp NOT NULL
);

-- sold — продан, refunded — возвращён, reserved — бронь до reserved_until
CREATE TABLE tickets (
  screening_id   int NOT NULL REFERENCES screenings(id),
  seat_id        int NOT NULL REFERENCES seats(id),
  status         text NOT NULL,
  reserved_until timestamp
);

INSERT INTO halls VALUES (1, 'Малый'), (2, 'Большой');

-- id места = зал * 100 + ряд * 10 + место
INSERT INTO seats (id, hall_id, row_no, seat_no)
SELECT 100 + r * 10 + s, 1, r, s
FROM generate_series(1, 2) r, generate_series(1, 6) s;
INSERT INTO seats (id, hall_id, row_no, seat_no)
SELECT 130 + s, 1, 3, s
FROM unnest(ARRAY[1, 2, 3, 5, 6, 7]) s;
INSERT INTO seats (id, hall_id, row_no, seat_no)
SELECT 200 + r * 10 + s, 2, r, s
FROM generate_series(1, 2) r, generate_series(1, 8) s;
UPDATE seats SET is_broken = true WHERE id = 113;

INSERT INTO screenings VALUES
  (1, 1, 'Дюна',        '2024-08-01 19:00'),
  (2, 1, 'Дюна',        '2024-08-01 22:00'),
  (3, 2, 'Оппенгеймер', '2024-08-01 20:00'),
  (4, 2, 'Оппенгеймер', '2024-08-02 20:00');

INSERT INTO tickets VALUES
  (1, 112, 'sold',     NULL),
  (1, 115, 'sold',     NULL),
  (1, 121, 'sold',     NULL),
  (1, 125, 'refunded', NULL),
  (1, 126, 'reserved', '2024-08-01 17:50'),
  (2, 123, 'sold',     NULL),
  (2, 124, 'sold',     NULL),
  (2, 121, 'reserved', '2024-08-01 18:30'),
  (3, 211, 'sold', NULL), (3, 212, 'sold', NULL), (3, 213, 'sold', NULL),
  (3, 214, 'sold', NULL), (3, 216, 'sold', NULL), (3, 217, 'sold', NULL),
  (3, 218, 'sold', NULL),
  (3, 221, 'sold', NULL), (3, 222, 'sold', NULL);
