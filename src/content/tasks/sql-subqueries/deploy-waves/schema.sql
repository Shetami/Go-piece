CREATE TABLE services (
  name text PRIMARY KEY
);

-- service нельзя выкатывать раньше, чем выкачен depends_on
CREATE TABLE deps (
  service    text NOT NULL REFERENCES services (name),
  depends_on text NOT NULL REFERENCES services (name),
  PRIMARY KEY (service, depends_on)
);

INSERT INTO services VALUES
  ('api'), ('auth'), ('billing'), ('cache'), ('cron'), ('db'),
  ('ledger'), ('notifier'), ('queue'), ('reports'), ('search');

INSERT INTO deps VALUES
  ('api',      'auth'),
  ('api',      'billing'),
  ('auth',     'db'),
  ('auth',     'cache'),
  ('billing',  'db'),
  ('billing',  'queue'),
  ('queue',    'db'),
  ('reports',  'billing'),
  ('reports',  'ledger'),
  ('ledger',   'reports'),
  ('notifier', 'ledger'),
  ('cron',     'cron');
