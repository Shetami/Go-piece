CREATE TABLE orders (
  id         int PRIMARY KEY,
  city       text,
  status     text,
  amount     numeric(10,2) NOT NULL,
  created_at timestamp NOT NULL
);

-- Сохранённые фильтры из админки. Пустое поле формы приходит как NULL
-- или как пустая строка — зависит от версии фронтенда.
CREATE TABLE saved_filters (
  name       text PRIMARY KEY,
  city       text,
  status     text,
  min_amount numeric(10,2),
  date_from  date,           -- включительно
  date_to    date            -- включительно
);

INSERT INTO orders VALUES
  (1, 'Москва',          'paid', 1500.00, '2024-02-29 23:30'),
  (2, 'москва ',         'new',   500.00, '2024-03-01 00:00'),
  (3, 'Санкт-Петербург', 'Paid', 1000.00, '2024-03-15 12:00'),
  (4, NULL,              'paid', 2500.00, '2024-03-20 10:00'),
  (5, 'Москва',          NULL,   1200.00, '2024-03-31 23:59'),
  (6, 'Москва',          'paid',  999.99, '2024-04-01 00:00'),
  (7, 'Казань',          'PAID', 3000.00, '2024-01-10 09:00');

INSERT INTO saved_filters VALUES
  ('moscow-march', ' Москва', NULL,   NULL,    '2024-03-01', '2024-03-31'),
  ('big-paid',     '',        'PAID', 1000.00, NULL,         NULL),
  ('till-feb',     NULL,      '  ',   NULL,    NULL,         '2024-02-29');
