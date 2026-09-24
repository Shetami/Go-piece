// @ts-check
import { defineConfig } from 'astro/config'
import mdx from '@astrojs/mdx'
import react from '@astrojs/react'
import vercel from '@astrojs/vercel'
import node from '@astrojs/node'
import tailwindcss from '@tailwindcss/vite'

/**
 * Сборка и дев-сервер по умолчанию делят один кэш пребандла — node_modules/.vite.
 * Если собрать проект, пока рядом работает `astro dev`, кэш под ним подменяется:
 * в браузере оказываются модули с разными версиями зависимостей, React грузится
 * дважды, и все острова падают на гидратации («Invalid hook call»). Разводим кэши.
 */
const isBuild = process.argv[2] === 'build' || process.env.npm_lifecycle_event === 'build'

/**
 * Куда собираем. По умолчанию — Vercel. С `DEPLOY_TARGET=node` получается
 * самостоятельный Node-сервер в dist/ (dist/server/entry.mjs + dist/client) —
 * его запускает Docker-образ.
 */
const target = process.env.DEPLOY_TARGET === 'node' ? 'node' : 'vercel'

/**
 * Сайт остаётся статическим: все страницы собираются заранее. Адаптер нужен
 * ровно ради маршрутов в src/pages/api — они помечены `prerender = false`
 * и живут на сервере, потому что компилировать Go в браузере нечем. SQL-задачам
 * сервер не нужен: Postgres для них работает прямо в браузере.
 */
export default defineConfig({
  site: 'https://go-piece.vercel.app',
  trailingSlash: 'ignore',
  adapter: target === 'node' ? node({ mode: 'standalone' }) : vercel(),
  integrations: [mdx(), react()],
  vite: {
    plugins: [tailwindcss()],
    cacheDir: isBuild ? 'node_modules/.vite-build' : 'node_modules/.vite',
    // PGlite (Postgres для SQL-задач) сам находит свои .wasm и .data рядом с
    // модулем через import.meta.url; пребандл это ломает, поэтому его не трогаем.
    optimizeDeps: { exclude: ['@electric-sql/pglite'] },
    // Воркер с Postgres импортирует модули — нужен ES-формат, а не iife.
    worker: { format: 'es' },
  },
})
