CREATE TABLE branches (
  id   int PRIMARY KEY,
  name text NOT NULL
);

-- Номер счёта сквозной внутри филиала и года, с 1 января начинается с 1.
-- Аннулированный счёт номер не освобождает.
CREATE TABLE invoices (
  id        int PRIMARY KEY,
  branch_id int NOT NULL REFERENCES branches(id),
  number    int NOT NULL,
  issued_on date NOT NULL,
  status    text NOT NULL -- issued, void
);

INSERT INTO branches VALUES (1, 'Север'), (2, 'Юг'), (3, 'Запад');

INSERT INTO invoices VALUES
  (1,  1, 1, '2023-11-01', 'issued'),
  (2,  1, 2, '2023-11-03', 'issued'),
  (3,  1, 3, '2023-11-10', 'void'),
  (4,  1, 5, '2023-11-15', 'issued'),
  (5,  1, 6, '2023-12-01', 'issued'),
  (6,  1, 9, '2023-12-31', 'issued'),
  (7,  1, 1, '2024-01-01', 'issued'),
  (8,  1, 2, '2024-01-09', 'issued'),
  (9,  2, 3, '2024-02-01', 'issued'),
  (10, 2, 4, '2024-02-02', 'void'),
  (11, 2, 4, '2024-02-02', 'issued'),
  (12, 2, 5, '2024-02-05', 'issued'),
  (13, 2, 8, '2024-02-10', 'issued'),
  (14, 3, 1, '2023-12-31', 'issued'),
  (15, 3, 1, '2024-01-02', 'issued'),
  (16, 3, 2, '2024-01-03', 'void'),
  (17, 3, 3, '2024-01-04', 'issued');
