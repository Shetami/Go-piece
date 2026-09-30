CREATE TABLE flights (
  id    int PRIMARY KEY,
  src   text NOT NULL,
  dst   text NOT NULL,
  price int  NOT NULL
);

INSERT INTO flights VALUES
  (1,  'MOW', 'LED', 3000),
  (2,  'LED', 'MOW', 2500),
  (3,  'MOW', 'KZN', 4000),
  (4,  'LED', 'KZN', 1000),
  (5,  'KZN', 'SVX', 2000),
  (6,  'MOW', 'SVX', 9000),
  (7,  'SVX', 'OVB', 3000),
  (8,  'LED', 'SVX', 3000),
  (9,  'OVB', 'MOW', 1000),
  (10, 'OVB', 'VVO', 7000),
  (11, 'MOW', 'AER', 5000),
  (12, 'KZN', 'AER', 2500),
  (13, 'SVX', 'KGD', 1000),
  (14, 'KGD', 'AER', 500),
  (15, 'MOW', 'MMK', 8000),
  (16, 'LED', 'MMK', 5000),
  (17, 'TJM', 'MOW', 4000);
