// @ts-check
import { defineConfig } from 'astro/config'
import mdx from '@astrojs/mdx'
import react from '@astrojs/react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  site: 'https://go-piece.vercel.app',
  trailingSlash: 'ignore',
  integrations: [mdx(), react()],
  vite: {
    plugins: [tailwindcss()],
  },
})
