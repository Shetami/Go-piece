-- События приходят с задержкой, поэтому id не отражает порядок во времени.
-- При равном времени события более позднее — то, у которого больше id.
CREATE TABLE parcel_events (
  id          int PRIMARY KEY,
  parcel_id   int NOT NULL,
  status      text NOT NULL,
  happened_at timestamp NOT NULL
);

INSERT INTO parcel_events VALUES
  (1,  1, 'created',    '2024-06-01 10:00'),
  (2,  1, 'in_transit', '2024-06-02 10:00'),
  (3,  1, 'delivered',  '2024-06-03 10:00'),
  (4,  2, 'created',    '2024-06-01 11:00'),
  (5,  2, 'in_transit', '2024-06-02 12:00'),
  (7,  2, 'returned',   '2024-06-02 12:00'),
  (6,  2, 'in_transit', '2024-06-02 11:00'),
  (8,  3, 'created',    '2024-06-05 09:00'),
  (9,  3, 'in_transit', '2024-06-06 09:00'),
  (10, 3, 'created',    '2024-06-05 08:00'),
  (11, 4, 'created',    '2024-06-07 10:00'),
  (12, 5, 'created',    '2024-06-01 09:00'),
  (13, 5, 'delivered',  '2024-06-08 18:00');
