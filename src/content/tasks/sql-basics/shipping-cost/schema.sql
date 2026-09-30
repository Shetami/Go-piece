CREATE TABLE tariffs (
  zone   int PRIMARY KEY,
  base   int NOT NULL,   -- ₽ за первый килограмм
  per_kg int NOT NULL    -- ₽ за каждый начатый килограмм сверх первого
);

CREATE TABLE parcels (
  id        int PRIMARY KEY,
  weight_kg numeric(5,2),  -- NULL: не взвешивали
  length_cm int,           -- габариты: все три или ни одного
  width_cm  int,
  height_cm int,
  zone      int            -- NULL: адрес не распознан
);

INSERT INTO tariffs VALUES
  (1, 300, 50),
  (2, 450, 80),
  (3, 600, 120);

INSERT INTO parcels VALUES
  (1, 0.80, 20, 15, 10, 1),
  (2, 0.50, 30, 20, 10, 1),
  (3, 2.00, NULL, NULL, NULL, 2),
  (4, NULL, 40, 30, 25, 1),
  (5, 1.00, 10, 10, 10, 1),
  (6, 3.30, 10, 10, 10, NULL),
  (7, NULL, NULL, NULL, NULL, 2),
  (8, 4.99, 50, 40, 30, 2);
