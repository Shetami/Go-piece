import { useEffect, useMemo, useRef, useState } from 'react'
import { Simulation } from '../engine/simulation.ts'
import type { ChanEvent, ChanEventType, ChanScenario } from '../engine/types.ts'

/** Прогоны длиннее этого стенд обрезает: история хранит полный снимок на каждый тик. */
const HARD_CAP = 400

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
  eventsUpTo: ChanEvent[]
  counts: Partial<Record<ChanEventType, number>>
  /** Событие, на котором автопауза остановилась в последний раз. */
  pausedOn: ChanEvent | null
}

/**
 * Прогоняет сценарий до конца сразу, а дальше UI просто двигает курсор по истории.
 * Движок детерминирован, поэтому «шаг назад» и перемотка — это чтение готового снимка.
 */
export function useSimulation(scenario: ChanScenario, seed: number, opts: { autoPause: boolean }): Playback {
  const sim = useMemo(
    () => new Simulation(scenario, seed).runToEnd(Math.min(scenario.stopAfter ?? 300, HARD_CAP)),
    [scenario, seed],
  )
  const last = sim.history.length - 1

  const [rawCursor, setCursorRaw] = useState(0)
  // Новый прогон может оказаться короче предыдущего, а сброс курсора произойдёт
  // только в эффекте — до него рендер обязан оставаться в пределах истории.
  const cursor = Math.min(rawCursor, last)
  const [playing, setPlaying] = useState(false)
  const [speed, setSpeed] = useState(4)
  const [pausedOn, setPausedOn] = useState<ChanEvent | null>(null)

  const seen = useRef(new Set<ChanEventType>())

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
    const hit = sim.history[cursor]?.events.find(
      (e) => sim.scenario.watchFor.includes(e.type) && !seen.current.has(e.type),
    )
    if (hit) {
      seen.current.add(hit.type)
      setPlaying(false)
      setPausedOn(hit)
    }
  }, [cursor, playing, last, sim])

  const eventsUpTo = useMemo(
    () => sim.history.slice(1, cursor + 1).flatMap((s) => s.events),
    [sim, cursor],
  )

  const counts = useMemo(() => {
    const c: Partial<Record<ChanEventType, number>> = {}
    for (const e of eventsUpTo) c[e.type] = (c[e.type] ?? 0) + 1
    return c
  }, [eventsUpTo])

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
    counts,
    pausedOn,
  }
}
