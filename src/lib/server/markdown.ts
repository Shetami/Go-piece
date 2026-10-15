/**
 * Markdown сгенерированных задач → HTML.
 *
 * Текст пишет модель, поэтому сырой HTML из него не пропускаем: он выводится
 * как текст. Ссылки — только http(s), без javascript: и прочего.
 *
 * Только сервер: страница рендерится на запрос, на клиент marked не уезжает.
 */
import { Marked } from 'marked'

const escape = (s: string) =>
  s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;')

const marked = new Marked({
  gfm: true,
  renderer: {
    html({ text }) {
      return escape(text)
    },
    link({ href, tokens }) {
      const label = this.parser.parseInline(tokens)
      if (!/^https?:\/\//i.test(href)) return label
      return `<a href="${escape(href)}" rel="nofollow noopener" target="_blank">${label}</a>`
    },
    image({ text }) {
      // Картинок в задачах не бывает, а чужой адрес в <img> — это запрос с браузера человека.
      return escape(text)
    },
  },
})

export function renderMarkdown(source: string): string {
  return marked.parse(source, { async: false })
}
