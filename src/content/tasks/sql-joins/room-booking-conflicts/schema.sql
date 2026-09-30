CREATE TABLE rooms (
  id   int PRIMARY KEY,
  name text NOT NULL
);

-- Бронь занимает переговорку на [starts_at, ends_at)
CREATE TABLE bookings (
  id        int PRIMARY KEY,
  room_id   int NOT NULL REFERENCES rooms (id),
  starts_at timestamp NOT NULL,
  ends_at   timestamp NOT NULL,
  status    text NOT NULL  -- 'confirmed', 'tentative', 'cancelled'
);

INSERT INTO rooms VALUES
  (1, 'Байкал'),
  (2, 'Ладога'),
  (3, 'Онега');

INSERT INTO bookings VALUES
  (1,  1, '2024-05-20 10:00', '2024-05-20 11:00', 'confirmed'),
  (2,  1, '2024-05-20 11:00', '2024-05-20 12:00', 'confirmed'),
  (3,  1, '2024-05-20 10:30', '2024-05-20 11:15', 'tentative'),
  (4,  1, '2024-05-20 09:00', '2024-05-20 13:00', 'cancelled'),
  (5,  2, '2024-05-20 09:00', '2024-05-20 18:00', 'confirmed'),
  (6,  2, '2024-05-20 12:00', '2024-05-20 12:30', 'confirmed'),
  (7,  3, '2024-05-20 10:00', '2024-05-20 11:00', 'confirmed'),
  (8,  3, '2024-05-20 11:00', '2024-05-20 11:30', 'confirmed'),
  (9,  2, '2024-05-21 09:00', '2024-05-21 10:00', 'confirmed'),
  (10, 1, '2024-05-20 10:30', '2024-05-20 11:15', 'confirmed');
