CREATE TABLE users (
  id    int PRIMARY KEY,
  email text NOT NULL
);

INSERT INTO users VALUES
  (1, 'anya@example.com'),
  (2, 'borya@example.com'),
  (3, 'Anya@Example.com'),
  (4, 'vera@example.com'),
  (5, 'borya@example.com'),
  (6, 'gosha@example.com'),
  (7, 'ANYA@EXAMPLE.COM'),
  (8, 'Vera@example.org');
