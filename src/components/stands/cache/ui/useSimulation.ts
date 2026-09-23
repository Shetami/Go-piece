import { useEffect, useMemo, useRef, useState } from 'react'
import { Simulation } from '../engine/simulation.ts'
import type { CacheEvent, CacheEventType, CacheScenario } from '../engine/types.ts'

/** Прогоны длиннее этого стенд обрезает: история хранит полный снимок на каждый тик. */
const HARD_CAP = 300

export interface Playback {
  sim: Simulation
  last: number
  cursor: number
  setCursor: (t: number) => void
  playing: boolean
  setPlaying: (v: boolean) => void
  speed: number
  setSpeed: (v: number) => void
  /** События, случившиеся на тиках 1..cursor. */
  eventsUpTo: CacheEvent[]
  /** Событие, на котором автопауза остановилась в последний раз. */
  pausedOn: CacheEvent | null
}

/**
 * Прогоняет сценарий до конца сразу, а дальше UI просто двигает курсор по истории.
 * Движок детерминирован, поэтому «шаг назад» и перемотка — это чтение готового кадра.
 */
export function useSimulation(scenario: CacheScenario, opts: { autoPause: boolean }): Playback {
  const sim = useMemo(() => new Simulation(scenario).runToEnd(Math.min(scenario.stopAfter, HARD_CAP)), [scenario])
  const last = sim.history.length - 1

  const [rawCursor, setCursorRaw] = useState(0)
  // Новый прогон может оказаться короче предыдущего, а сброс курсора произойдёт
  // только в эффекте — до него рендер обязан оставаться в пределах истории.
  const cursor = Math.min(rawCursor, last)
  const [playing, setPlaying] = useState(false)
  const [speed, setSpeed] = useState(5)
  const [pausedOn, setPausedOn] = useState<CacheEvent | null>(null)

  const seen = useRef(new Set<CacheEventType>())

  useEffect(() => {
    setCursorRaw(0)
    setPlaying(false)
    setPausedOn(null)
    seen.current = new Set()
  }, [sim])

  const setCursor = (t: number) => setCursorRaw(Math.max(0, Math.min(last, t)))

  const autoPause = useRef(opts.autoPause)
  autoPause.current = opts.autoPause

  useEffect(() => {
    if (!playing) return
    const id = window.setInterval(() => setCursorRaw((c) => Math.min(c + 1, last)), 1000 / speed)
    return () => window.clearInterval(id)
  }, [playing, speed, last])

  useEffect(() => {
    if (!playing) return
    if (cursor >= last) {
      setPlaying(false)
      return
    }
    if (!autoPause.current) return
    const hit = sim.history[cursor]?.events.find((e) => sim.scenario.watchFor.includes(e.type) && !seen.current.has(e.type))
    if (hit) {
      seen.current.add(hit.type)
      setPlaying(false)
      setPausedOn(hit)
    }
  }, [cursor, playing, last, sim])

  const eventsUpTo = useMemo(() => sim.history.slice(1, cursor + 1).flatMap((f) => f.events), [sim, cursor])

  return {
    sim,
    last,
    cursor,
    setCursor,
    playing,
    setPlaying: (v) => {
      if (v && cursor >= last) setCursorRaw(0)
      if (v) setPausedOn(null)
      setPlaying(v)
    },
    speed,
    setSpeed,
    eventsUpTo,
    pausedOn,
  }
}
