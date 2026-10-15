/**
 * База аккаунтов — SQLite из стандартной библиотеки Node (node:sqlite), без
 * нативных зависимостей. Один файл: в Docker он лежит в volume /app/data,
 * локально — в ./data. Путь меняется переменной DATABASE_PATH.
 *
 * Только сервер. Соединение одно на процесс и открывается при первом запросе.
 */
import { mkdirSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { DatabaseSync } from 'node:sqlite'

const SCHEMA = `
  CREATE TABLE IF NOT EXISTS users (
    id            INTEGER PRIMARY KEY,
    email         TEXT    NOT NULL UNIQUE,
    password_hash TEXT    NOT NULL,
    created_at    INTEGER NOT NULL
  );

  -- В базе лежит не сам токен сессии, а его SHA-256: утёкшая база не даёт войти.
  CREATE TABLE IF NOT EXISTS sessions (
    token_hash TEXT    PRIMARY KEY,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at INTEGER NOT NULL
  );
  CREATE INDEX IF NOT EXISTS sessions_user ON sessions(user_id);

  -- Прогресс — ключи localStorage как есть, см. src/lib/sync.ts.
  CREATE TABLE IF NOT EXISTS user_state (
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key        TEXT    NOT NULL,
    value      TEXT    NOT NULL,
    updated_at INTEGER NOT NULL,
    PRIMARY KEY (user_id, key)
  );

  -- Задачи, сгенерированные на сайте, см. src/lib/server/generated.ts. Скрытые
  -- части (тесты, эталонный вывод, строка с багом) лежат здесь же и наружу
  -- отдаются только маршрутом /api/check.
  CREATE TABLE IF NOT EXISTS generated_tasks (
    id          TEXT    PRIMARY KEY,
    user_id     INTEGER REFERENCES users(id) ON DELETE SET NULL,
    topic       TEXT    NOT NULL,
    kind        TEXT    NOT NULL,
    level       TEXT    NOT NULL,
    title       TEXT    NOT NULL,
    description TEXT    NOT NULL,
    statement   TEXT    NOT NULL,
    explanation TEXT    NOT NULL,
    starter     TEXT    NOT NULL,
    solution    TEXT    NOT NULL,
    tests       TEXT,
    expect      TEXT,
    bug_line    INTEGER,
    created_at  INTEGER NOT NULL
  );
  CREATE INDEX IF NOT EXISTS generated_tasks_user ON generated_tasks(user_id, created_at);

  -- Каждая попытка генерации, удачная или нет: по ним считается суточный лимит.
  CREATE TABLE IF NOT EXISTS generation_attempts (
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at INTEGER NOT NULL
  );
  CREATE INDEX IF NOT EXISTS generation_attempts_user ON generation_attempts(user_id, created_at);
`

let db: DatabaseSync | null = null

export function getDb(): DatabaseSync {
  if (db) return db
  const path = resolve(process.env.DATABASE_PATH ?? 'data/go-piece.db')
  mkdirSync(dirname(path), { recursive: true })
  db = new DatabaseSync(path)
  db.exec('PRAGMA journal_mode = WAL; PRAGMA foreign_keys = ON; PRAGMA busy_timeout = 5000;')
  db.exec(SCHEMA)
  return db
}

/** Несколько запросов одной транзакцией. node:sqlite синхронный, так что fn тоже синхронна. */
export function transaction<T>(fn: (db: DatabaseSync) => T): T {
  const d = getDb()
  d.exec('BEGIN IMMEDIATE')
  try {
    const result = fn(d)
    d.exec('COMMIT')
    return result
  } catch (err) {
    d.exec('ROLLBACK')
    throw err
  }
}
