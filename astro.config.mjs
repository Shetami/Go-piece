// @ts-check
import { defineConfig } from 'astro/config'
import mdx from '@astrojs/mdx'
import react from '@astrojs/react'
import tailwindcss from '@tailwindcss/vite'

/**
 * Сборка и дев-сервер по умолчанию делят один кэш пребандла — node_modules/.vite.
 * Если собрать проект, пока рядом работает `astro dev`, кэш под ним подменяется:
 * в браузере оказываются модули с разными версиями зависимостей, React грузится
 * дважды, и все острова падают на гидратации («Invalid hook call»). Разводим кэши.
 */
const isBuild = process.argv[2] === 'build' || process.env.npm_lifecycle_event === 'build'

export default defineConfig({
  site: 'https://go-piece.vercel.app',
  trailingSlash: 'ignore',
  integrations: [mdx(), react()],
  vite: {
    plugins: [tailwindcss()],
    cacheDir: isBuild ? 'node_modules/.vite-build' : 'node_modules/.vite',
  },
})
