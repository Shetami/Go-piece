CREATE TABLE prices (
  ticker text NOT NULL,
  day    date NOT NULL,
  price  numeric(10, 2) NOT NULL,
  PRIMARY KEY (ticker, day)
);

INSERT INTO prices VALUES
  ('AAA', '2024-05-01', 10.00),
  ('AAA', '2024-05-02', 10.50),
  ('AAA', '2024-05-03',  9.80),
  ('BBB', '2024-05-01', 50.00),
  ('BBB', '2024-05-02', 52.25);
