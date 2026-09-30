CREATE TABLE departments (
  id   int PRIMARY KEY,
  name text NOT NULL
);

-- department_id пуст у стажёров, которых ещё не распределили
CREATE TABLE managers (
  id            int PRIMARY KEY,
  name          text NOT NULL,
  department_id int REFERENCES departments (id)
);

-- manager_id пуст у сделок из входящего потока, которые ещё никто не взял
CREATE TABLE deals (
  id         int PRIMARY KEY,
  client     text NOT NULL,
  status     text NOT NULL,  -- 'open', 'won', 'lost'
  manager_id int REFERENCES managers (id)
);

CREATE TABLE calls (
  id        int PRIMARY KEY,
  deal_id   int NOT NULL REFERENCES deals (id),
  called_at timestamp NOT NULL
);

INSERT INTO departments VALUES
  (1, 'Корпоративные'),
  (2, 'Малый бизнес');

INSERT INTO managers VALUES
  (10, 'Ольга', 1),
  (11, 'Павел', 2),
  (12, 'Рома',  NULL);

INSERT INTO deals VALUES
  (100, 'ООО Ромашка',  'open', 10),
  (101, 'ИП Сидоров',   'open', 11),
  (102, 'АО Вектор',    'open', 12),
  (103, 'ООО Лютик',    'open', NULL),
  (104, 'ЗАО Берёзка',  'won',  10),
  (105, 'ООО Нарцисс',  'open', 11);

INSERT INTO calls VALUES
  (1, 100, '2024-06-24 10:00'),
  (2, 100, '2024-06-28 16:30'),
  (3, 100, '2024-06-20 12:00'),
  (4, 101, '2024-06-30 23:10'),
  (5, 102, '2024-07-01 00:00'),
  (6, 103, '2024-06-25 09:00'),
  (7, 104, '2024-06-26 11:00');
