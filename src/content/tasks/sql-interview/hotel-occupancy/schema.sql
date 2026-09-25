CREATE TABLE room_types (
  id   int PRIMARY KEY,
  name text NOT NULL
);

CREATE TABLE rooms (
  id           int PRIMARY KEY, -- номер комнаты
  room_type_id int NOT NULL REFERENCES room_types(id)
);

CREATE TABLE guests (
  id           int PRIMARY KEY,
  name         text NOT NULL,
  discount_pct int NOT NULL DEFAULT 0 -- персональная скидка на проживание
);

-- Гость живёт ночи с check_in по check_out − 1; в день выезда номер свободен
CREATE TABLE bookings (
  id        int PRIMARY KEY,
  room_id   int NOT NULL REFERENCES rooms(id),
  guest_id  int NOT NULL REFERENCES guests(id),
  check_in  date NOT NULL,
  check_out date NOT NULL,
  status    text NOT NULL -- confirmed, cancelled
);

-- Цена ночи зависит от типа номера и даты; valid_to включительно
CREATE TABLE tariffs (
  room_type_id    int NOT NULL REFERENCES room_types(id),
  valid_from      date NOT NULL,
  valid_to        date NOT NULL,
  price_per_night int NOT NULL
);

INSERT INTO room_types VALUES
  (1, 'Стандарт'),
  (2, 'Комфорт'),
  (3, 'Люкс');

INSERT INTO rooms VALUES
  (101, 1), (102, 1), (103, 1),
  (201, 2), (202, 2),
  (301, 3);

INSERT INTO guests VALUES
  (1, 'Аня',  0),
  (2, 'Боря', 10),
  (3, 'Вера', 0),
  (4, 'Гоша', 20);

INSERT INTO tariffs VALUES
  (1, '2024-01-01', '2024-03-06', 3000),
  (1, '2024-03-07', '2024-03-10', 4000),
  (1, '2024-03-11', '2024-12-31', 3000),
  (2, '2024-01-01', '2024-03-15', 5000),
  (2, '2024-03-16', '2024-12-31', 6000),
  (3, '2024-01-01', '2024-12-31', 10000);

INSERT INTO bookings VALUES
  (1, 101, 1, '2024-02-27', '2024-03-03', 'confirmed'),
  (2, 102, 2, '2024-03-05', '2024-03-09', 'confirmed'),
  (3, 103, 3, '2024-03-20', '2024-03-22', 'cancelled'),
  (4, 101, 4, '2024-03-30', '2024-04-02', 'confirmed'),
  (5, 201, 3, '2024-03-14', '2024-03-18', 'confirmed'),
  (6, 202, 1, '2024-03-01', '2024-03-02', 'confirmed'),
  (7, 301, 2, '2024-04-05', '2024-04-07', 'confirmed');
