CREATE TABLE accounts (
  id        int PRIMARY KEY,
  name      text    NOT NULL,
  is_active boolean NOT NULL
);

-- Контактов у клиента несколько, основной помечен is_primary
CREATE TABLE contacts (
  id         int PRIMARY KEY,
  account_id int     NOT NULL REFERENCES accounts (id),
  email      text    NOT NULL,
  is_primary boolean NOT NULL
);

CREATE TABLE invoices (
  id         int PRIMARY KEY,
  account_id int     NOT NULL REFERENCES accounts (id),
  issued_on  date    NOT NULL,
  amount     numeric NOT NULL
);

-- Платежи привязаны к клиенту, а не к счёту (аванс, частичная оплата)
CREATE TABLE payments (
  id         int PRIMARY KEY,
  account_id int     NOT NULL REFERENCES accounts (id),
  paid_on    date    NOT NULL,
  amount     numeric NOT NULL
);

CREATE TABLE tickets (
  id         int PRIMARY KEY,
  account_id int  NOT NULL REFERENCES accounts (id),
  status     text NOT NULL  -- 'open', 'closed'
);

INSERT INTO accounts VALUES
  (1, 'Альфа',   true),
  (2, 'Бета',    true),
  (3, 'Гамма',   true),
  (4, 'Дельта',  false),
  (5, 'Эпсилон', true);

INSERT INTO contacts VALUES
  (1, 1, 'ceo@alfa.example',   true),
  (2, 1, 'buh@alfa.example',   false),
  (3, 2, 'info@beta.example',  false),
  (4, 3, 'it@gamma.example',   true),
  (5, 4, 'boss@delta.example', true);

INSERT INTO invoices VALUES
  (1, 1, '2024-01-10', 1000),
  (2, 1, '2024-02-10', 1000),
  (3, 1, '2023-12-10', 700),
  (4, 2, '2024-03-01', 500),
  (5, 4, '2024-01-15', 900);

INSERT INTO payments VALUES
  (1, 1, '2024-01-12', 600),
  (2, 1, '2024-01-20', 400),
  (3, 1, '2024-02-15', 500),
  (4, 1, '2023-12-20', 700),
  (5, 3, '2024-02-01', 300),
  (6, 4, '2024-01-20', 900);

INSERT INTO tickets VALUES
  (1, 1, 'open'),
  (2, 1, 'closed'),
  (3, 3, 'open'),
  (4, 3, 'open'),
  (5, 4, 'open');
