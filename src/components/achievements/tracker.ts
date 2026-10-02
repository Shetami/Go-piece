/**
 * Трекер достижений — подключён в BaseLayout, работает на каждой странице.
 *
 * Сам ничего не решает про достижения: отмечает события по разметке страницы
 * (дочитана лекция, тронут стенд, открыт термин, 404), а по каждому
 * PROGRESS_EVENT пересчитывает достижения и показывает уведомление о новых.
 *
 * Разметка, на которую он смотрит:
 *   [data-lecture-end="<id>"]  конец текста лекции: доскроллили — лекция прочитана
 *   [data-stand="<id>"]        обёртка стенда: клик или клавиша внутри — стенд тронут
 *   [data-term="<id>"]         страница термина справочника
 *   [data-lost]                страница 404
 */
import { evaluate, formatBerry, type Catalog, type Evaluated } from '../../lib/achievements.ts'
import {
  PROGRESS_EVENT,
  loadEarned,
  loadProgress,
  recordFlag,
  recordRead,
  recordStand,
  recordTerm,
  saveEarned,
} from '../../lib/progress.ts'

/** Больше стольких уведомлений за раз не показываем — сворачиваем в одно. */
const MAX_TOASTS = 3

let catalog: Promise<Catalog | null> | null = null

function getCatalog(): Promise<Catalog | null> {
  catalog ??= fetch('/achievements/catalog.json')
    .then((r) => (r.ok ? (r.json() as Promise<Catalog>) : null))
    .catch(() => null)
  return catalog
}

async function check(): Promise<void> {
  const c = await getCatalog()
  if (!c) return
  const earned = loadEarned()
  const results = evaluate(loadProgress(c), c, earned)
  const fresh = results.filter((e) => e.unlocked && e.unlockedAt === undefined)
  if (fresh.length === 0) return
  const now = Date.now()
  for (const e of fresh) earned[e.achievement.id] = now
  saveEarned(earned)
  announce(fresh)
}

/** Проверки идут строго по очереди: две параллельные объявили бы одно достижение дважды. */
let queue = Promise.resolve()
function schedule(): void {
  queue = queue.then(check, check)
}

function toastRoot(): HTMLElement {
  let root = document.getElementById('ach-toasts')
  if (!root) {
    root = document.createElement('div')
    root.id = 'ach-toasts'
    root.className = 'ach-toasts'
    root.setAttribute('role', 'status')
    root.setAttribute('aria-live', 'polite')
    document.body.append(root)
  }
  return root
}

function toast(kicker: string, title: string, sub: string, href: string): void {
  const el = document.createElement('a')
  el.className = 'ach-toast'
  el.href = href
  const k = document.createElement('span')
  k.className = 'ach-toast-kicker'
  k.textContent = kicker
  const t = document.createElement('span')
  t.className = 'ach-toast-title'
  t.textContent = title
  const s = document.createElement('span')
  s.className = 'ach-toast-sub'
  s.textContent = sub
  el.append(k, t, s)
  toastRoot().append(el)
  window.setTimeout(() => el.classList.add('is-leaving'), 6000)
  window.setTimeout(() => el.remove(), 6600)
}

function announce(fresh: Evaluated[]): void {
  if (fresh.length > MAX_TOASTS) {
    const total = fresh.reduce((s, e) => s + e.achievement.bounty, 0)
    toast('Достижения получены', `Новых достижений: ${fresh.length}`, `+${formatBerry(total)} ฿ к награде`, '/achievements/')
    return
  }
  for (const e of fresh) {
    const a = e.achievement
    toast('Достижение получено', a.title, `+${formatBerry(a.bounty)} ฿ · ${a.task}`, `/achievements/#${a.id}`)
  }
}

function watchPage(): void {
  const end = document.querySelector<HTMLElement>('[data-lecture-end]')
  if (end?.dataset.lectureEnd && 'IntersectionObserver' in window) {
    const id = end.dataset.lectureEnd
    const io = new IntersectionObserver((entries) => {
      if (entries.some((e) => e.isIntersecting)) {
        io.disconnect()
        recordRead(id)
      }
    })
    io.observe(end)
  }

  const term = document.querySelector<HTMLElement>('[data-term]')?.dataset.term
  if (term) recordTerm(term)

  if (document.querySelector('[data-lost]')) recordFlag('lost')

  const touched = new Set<string>()
  const onStand = (ev: Event) => {
    const id = (ev.target as Element | null)?.closest?.<HTMLElement>('[data-stand]')?.dataset.stand
    if (!id || touched.has(id)) return
    touched.add(id)
    recordStand(id)
  }
  document.addEventListener('pointerdown', onStand, { capture: true })
  document.addEventListener('keydown', onStand, { capture: true })
}

window.addEventListener(PROGRESS_EVENT, schedule)
watchPage()
// Первая проверка — для прогресса, набранного до появления достижений.
schedule()
