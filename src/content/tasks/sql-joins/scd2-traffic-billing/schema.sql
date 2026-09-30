CREATE TABLE clients (
  id   int PRIMARY KEY,
  name text NOT NULL
);

-- Справочник тарифов. При смене цены старую строку не удаляют, а архивируют:
-- у одного code бывает несколько строк, действующая — archived = false.
CREATE TABLE tariffs (
  code        text    NOT NULL,
  price_per_gb numeric NOT NULL,
  archived    boolean NOT NULL
);

-- История тарифа клиента (SCD2): [valid_from, valid_to), valid_to пуст у текущей версии.
-- Менеджеры иногда заводят новую версию, не закрыв старую, — интервалы пересекаются.
-- Правило: действует версия с самым поздним valid_from, при равенстве — загруженная позже.
CREATE TABLE client_tariffs (
  client_id   int       NOT NULL REFERENCES clients (id),
  tariff_code text      NOT NULL,
  valid_from  date      NOT NULL,
  valid_to    date,
  loaded_at   timestamp NOT NULL
);

CREATE TABLE usage (
  client_id int     NOT NULL REFERENCES clients (id),
  day       date    NOT NULL,
  gb        numeric NOT NULL
);

INSERT INTO clients VALUES
  (1, 'Астра'),
  (2, 'Бриз'),
  (3, 'Вега'),
  (4, 'Гранит');

INSERT INTO tariffs VALUES
  ('start', 2.00, true),
  ('start', 2.50, false),
  ('pro',   1.20, false),
  ('max',   0.80, false);

INSERT INTO client_tariffs VALUES
  (1, 'start', '2024-01-01', '2024-06-15', '2024-01-01 09:00'),
  (1, 'pro',   '2024-06-15', NULL,         '2024-06-14 18:00'),
  (2, 'start', '2024-03-01', NULL,         '2024-03-01 09:00'),
  (2, 'max',   '2024-06-10', NULL,         '2024-06-10 09:00'),
  (2, 'pro',   '2024-06-10', '2024-06-20', '2024-06-10 12:00'),
  (3, 'pro',   '2024-06-05', NULL,         '2024-06-05 10:00');

INSERT INTO usage VALUES
  (1, '2024-06-14', 10),
  (1, '2024-06-15', 10),
  (1, '2024-05-31', 100),
  (2, '2024-06-09', 10),
  (2, '2024-06-10', 10),
  (2, '2024-06-19', 10),
  (2, '2024-06-20', 10),
  (3, '2024-06-01', 5),
  (3, '2024-06-05', 10),
  (3, '2024-06-30', 2.5);
