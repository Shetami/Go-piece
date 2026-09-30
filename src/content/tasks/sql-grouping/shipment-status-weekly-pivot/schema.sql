CREATE TABLE shipments (
  id         int PRIMARY KEY,
  created_at date NOT NULL,
  status     text          -- приходит от разных перевозчиков как есть
);

INSERT INTO shipments VALUES
  (1,  '2024-03-31', 'delivered'),
  (2,  '2024-04-01', 'delivered'),
  (3,  '2024-04-03', 'Delivered '),
  (4,  '2024-04-07', 'in_transit'),
  (5,  '2024-04-08', 'LOST'),
  (6,  '2024-04-09', NULL),
  (7,  '2024-04-10', 'returned'),
  (8,  '2024-04-22', ' In_Transit'),
  (9,  '2024-04-28', 'delivered'),
  (10, '2024-04-29', 'delivered'),
  (11, '2024-05-05', 'lost'),
  (12, '2024-05-06', 'delivered');
