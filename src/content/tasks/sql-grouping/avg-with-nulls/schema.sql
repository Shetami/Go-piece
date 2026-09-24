CREATE TABLE reviews (
  id     int PRIMARY KEY,
  rating int
);

-- rating пуст, если человек оставил только текст отзыва, без оценки
INSERT INTO reviews VALUES
  (1, 5),
  (2, 3),
  (3, NULL),
  (4, 4),
  (5, NULL),
  (6, 3);
