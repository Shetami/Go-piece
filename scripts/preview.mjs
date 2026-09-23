/**
 * Просмотр собранного сайта.
 *
 * `astro preview` с адаптером Vercel не работает: адаптер собирает не сервер,
 * который можно запустить, а раскладку `.vercel/output` для чужой платформы.
 * А посмотреть сборку надо — ровно ради поиска: индекс Pagefind появляется
 * только после `pnpm build`, и на дев-сервере его нет.
 *
 * Поэтому здесь просто отдаётся статика из `.vercel/output/static`. Маршрутов
 * `/api/*` тут нет и быть не может — они живут в серверной функции. Задачи
 * практики проверяются на `pnpm dev`, где эти маршруты работают по-настоящему.
 */

import { createServer } from 'node:http'
import { createReadStream } from 'node:fs'
import { stat } from 'node:fs/promises'
import { extname, join, normalize, resolve } from 'node:path'

const ROOT = resolve('.vercel/output/static')
const PORT = Number(process.env.PORT ?? 4322)

const TYPES = {
  '.html': 'text/html; charset=utf-8',
  '.js': 'text/javascript; charset=utf-8',
  '.mjs': 'text/javascript; charset=utf-8',
  '.css': 'text/css; charset=utf-8',
  '.json': 'application/json; charset=utf-8',
  '.svg': 'image/svg+xml',
  '.png': 'image/png',
  '.jpg': 'image/jpeg',
  '.webp': 'image/webp',
  '.ico': 'image/x-icon',
  '.woff2': 'font/woff2',
  '.wasm': 'application/wasm',
  '.pf_meta': 'application/octet-stream',
  '.pf_fragment': 'application/octet-stream',
  '.pf_index': 'application/octet-stream',
}

/** Путь из запроса → файл внутри ROOT, или null, если его пытаются увести наружу. */
async function resolveFile(urlPath) {
  const clean = normalize(decodeURIComponent(urlPath.split('?')[0])).replace(/^(\.\.[/\\])+/, '')
  const target = join(ROOT, clean)
  if (!target.startsWith(ROOT)) return null

  for (const candidate of [target, join(target, 'index.html'), `${target}.html`]) {
    try {
      if ((await stat(candidate)).isFile()) return candidate
    } catch {
      // Не тот вариант — пробуем следующий.
    }
  }
  return null
}

createServer(async (req, res) => {
  const file = (await resolveFile(req.url ?? '/')) ?? (await resolveFile('/404'))
  if (!file) {
    res.writeHead(404, { 'content-type': 'text/plain; charset=utf-8' })
    res.end('404')
    return
  }

  res.writeHead(file.endsWith('404.html') ? 404 : 200, {
    'content-type': TYPES[extname(file)] ?? 'application/octet-stream',
    'cache-control': 'no-store',
  })
  createReadStream(file).pipe(res)
}).listen(PORT, () => {
  console.log(`Собранный сайт: http://localhost:${PORT}`)
  console.log('Маршрутов /api здесь нет — для задач практики нужен pnpm dev.')
})
