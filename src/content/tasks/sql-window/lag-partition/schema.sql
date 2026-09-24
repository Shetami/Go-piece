CREATE TABLE daily_revenue (
  shop    text NOT NULL,
  day     date NOT NULL,
  revenue int NOT NULL,
  PRIMARY KEY (shop, day)
);

INSERT INTO daily_revenue VALUES
  ('Арбат',  '2024-05-01', 1000),
  ('Арбат',  '2024-05-02', 1200),
  ('Арбат',  '2024-05-03',  900),
  ('Тверская', '2024-05-01', 3000),
  ('Тверская', '2024-05-02', 3500),
  ('Тверская', '2024-05-03', 3100);
