/**
 * Остров страницы /achievements/: плакат «WANTED» с общей наградой и рангом,
 * под ним все достижения по группам с прогрессом.
 *
 * Каталог приходит пропсом со сборки, прогресс читается из localStorage после
 * гидратации — до неё (и на сервере) все достижения выглядят неполученными.
 * Даты получения записывает трекер (он работает и на этой странице) —
 * остров перечитывает их по EARNED_EVENT, а изменения из других вкладок — по storage.
 */
import { useEffect, useState } from 'react'
import { GROUPS, evaluate, formatBerry, rankFor, type Catalog, type Evaluated } from '../../lib/achievements.ts'
import { EARNED_EVENT, PROGRESS_EVENT, loadEarned, loadProgress } from '../../lib/progress.ts'

interface Props {
  catalog: Catalog
}

const EMPTY = {
  read: new Set<string>(),
  quizzes: new Map(),
  solved: new Set<string>(),
  stands: new Set<string>(),
  terms: new Set<string>(),
  days: [],
  bestDay: 0,
  flags: new Set<string>(),
}

function Card({ e }: { e: Evaluated }) {
  const a = e.achievement
  const hidden = a.secret && !e.unlocked
  const pct = e.need > 0 ? Math.round((e.have / e.need) * 100) : 0
  return (
    <li id={a.id} className={['ach-card', e.unlocked ? 'is-unlocked' : 'is-locked'].join(' ')}>
      <div className="ach-card-head">
        <span className="ach-seal" aria-hidden>
          {e.unlocked ? '☠' : hidden ? '?' : '⚓'}
        </span>
        <span className="ach-bounty">{formatBerry(a.bounty)} ฿</span>
      </div>
      <h3 className="ach-title">{hidden ? '???' : a.title}</h3>
      {!hidden && <p className="ach-quote">{a.quote}</p>}
      <p className="ach-task">{hidden ? 'Секретное достижение. Условие откроется, когда вы его получите.' : a.task}</p>
      {e.unlocked ? (
        <p className="ach-done">
          Получено{e.unlockedAt ? ` ${new Date(e.unlockedAt).toLocaleDateString('ru-RU')}` : ''}
        </p>
      ) : (
        !hidden &&
        e.need > 1 && (
          <div className="ach-progress" aria-label={`Прогресс: ${e.have} из ${e.need}`}>
            <div className="ach-bar">
              <div className="ach-bar-fill" style={{ width: `${pct}%` }} />
            </div>
            <span className="ach-count">
              {e.have} / {e.need}
            </span>
          </div>
        )
      )}
    </li>
  )
}

export default function Achievements({ catalog }: Props) {
  const [results, setResults] = useState<Evaluated[]>(() => evaluate(EMPTY, catalog))

  useEffect(() => {
    const refresh = () => setResults(evaluate(loadProgress(catalog), catalog, loadEarned()))
    refresh()
    const events = [PROGRESS_EVENT, EARNED_EVENT, 'storage']
    for (const ev of events) window.addEventListener(ev, refresh)
    return () => {
      for (const ev of events) window.removeEventListener(ev, refresh)
    }
  }, [catalog])

  const rank = rankFor(results)
  const got = results.filter((e) => e.unlocked).length
  const toNext = rank.next ? Math.min(100, Math.round((rank.bounty / rank.next.min) * 100)) : 100

  return (
    <div className="ach">
      <section className="ach-poster" aria-label="Ваша награда">
        <div className="ach-poster-wanted">WANTED</div>
        <img src="/logo.png" alt="" width="140" height="140" className="ach-poster-img" />
        <div className="ach-poster-rank">{rank.title}</div>
        <div className="ach-poster-dead">DEAD OR ALIVE</div>
        <div className="ach-poster-bounty">
          <span aria-hidden>฿</span> {formatBerry(rank.bounty)}
          <span className="ach-poster-dash">-</span>
        </div>
        <div className="ach-poster-marine">MARINE</div>
      </section>

      <div className="ach-summary">
        <p>
          Получено <b>{got}</b> из {results.length} достижений.
        </p>
        {rank.next && (
          <div className="ach-progress ach-progress-rank">
            <div className="ach-bar">
              <div className="ach-bar-fill" style={{ width: `${toNext}%` }} />
            </div>
            <span className="ach-count">
              до ранга «{rank.next.title}» — {formatBerry(Math.max(0, rank.next.min - rank.bounty))} ฿
            </span>
          </div>
        )}
        <p className="text-sm text-muted">
          Прогресс хранится только в этом браузере: в другом браузере или после очистки данных сайта награда начнётся
          с нуля.
        </p>
      </div>

      {GROUPS.map((g) => {
        const items = results.filter((e) => e.achievement.group === g.id)
        if (items.length === 0) return null
        return (
          <section key={g.id} className="ach-group">
            <h2 className="ach-group-title">
              {g.title}
              <span className="ach-group-count">
                {items.filter((e) => e.unlocked).length} / {items.length}
              </span>
            </h2>
            <p className="ach-group-blurb">{g.blurb}</p>
            <ul className="ach-grid">
              {items.map((e) => (
                <Card key={e.achievement.id} e={e} />
              ))}
            </ul>
          </section>
        )
      })}
    </div>
  )
}
