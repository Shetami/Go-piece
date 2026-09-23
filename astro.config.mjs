// @ts-check
import { defineConfig } from 'astro/config'
import mdx from '@astrojs/mdx'
import react from '@astrojs/react'
import vercel from '@astrojs/vercel'
import tailwindcss from '@tailwindcss/vite'

/**
 * Сборка и дев-сервер по умолчанию делят один кэш пребандла — node_modules/.vite.
 * Если собрать проект, пока рядом работает `astro dev`, кэш под ним подменяется:
 * в браузере оказываются модули с разными версиями зависимостей, React грузится
 * дважды, и все острова падают на гидратации («Invalid hook call»). Разводим кэши.
 */
const isBuild = process.argv[2] === 'build' || process.env.npm_lifecycle_event === 'build'

/**
 * Сайт остаётся статическим: все страницы собираются заранее. Адаптер нужен
 * ровно ради двух маршрутов в src/pages/api — они помечены `prerender = false`
 * и живут на сервере, потому что компилировать Go в браузере нечем.
 */
export default defineConfig({
  site: 'https://go-piece.vercel.app',
  trailingSlash: 'ignore',
  adapter: vercel(),
  integrations: [mdx(), react()],
  vite: {
    plugins: [tailwindcss()],
    cacheDir: isBuild ? 'node_modules/.vite-build' : 'node_modules/.vite',
  },
})
