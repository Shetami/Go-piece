CREATE TABLE shipments (
  id         int PRIMARY KEY,
  created_at timestamp NOT NULL
);

-- События приходят от разных систем и пишутся не по порядку.
CREATE TABLE shipment_events (
  id          int PRIMARY KEY,
  shipment_id int NOT NULL REFERENCES shipments(id),
  status      text NOT NULL,
  happened_at timestamp NOT NULL
);

INSERT INTO shipments VALUES
  (1,  '2024-06-25 09:00'),
  (2,  '2024-06-26 07:00'),
  (3,  '2024-06-29 11:00'),
  (4,  '2024-06-27 08:00'),
  (5,  '2024-06-26 17:00'),
  (6,  '2024-06-24 09:00'),
  (7,  '2024-06-28 12:00'),
  (8,  '2024-06-30 09:00'),
  (9,  '2024-06-20 09:00'),
  (10, '2024-06-27 10:00');

INSERT INTO shipment_events VALUES
  (1,  1,  'accepted',    '2024-06-25 10:00'),
  (2,  1,  'in_transit',  '2024-06-26 09:00'),
  (3,  1,  'delivered',   '2024-06-27 15:00'),
  (4,  2,  'accepted',    '2024-06-26 08:00'),
  (5,  2,  'in_transit',  '2024-06-28 07:30'),
  (6,  3,  'accepted',    '2024-06-29 12:00'),
  (7,  4,  'in_transit',  '2024-06-27 10:00'),
  (8,  4,  'Delivered ',  '2024-06-27 10:00'),
  (9,  5,  'accepted',    '2024-06-26 18:00'),
  (10, 5,  'in_transit',  '2024-06-26 18:00'),
  (11, 6,  'in_transit',  '2024-06-30 20:00'),
  (12, 6,  'accepted',    '2024-06-24 10:00'),
  (13, 9,  'CANCELLED',   '2024-06-20 10:00'),
  (14, 10, 'in_transit',  '2024-06-28 11:30');
