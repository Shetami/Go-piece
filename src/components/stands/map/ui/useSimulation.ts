import { useEffect, useMemo, useRef, useState } from 'react'
import { Simulation } from '../engine/simulation.ts'
import type { MapEvent, MapEventType, MapScenario } from '../engine/types.ts'

/** Прогоны длиннее этого стенд обрезает: история хранит полный снимок на каждый тик. */
export const HARD_CAP = 600

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
  eventsUpTo: MapEvent[]
  /** Событие, на котором автопауза остановилась в последний раз. */
  pausedOn: MapEvent | null
  /**
   * Следующий пересчёт прогона не сбрасывает курсор в начало, а продолжает с тика `at`:
   * так команда из консоли проигрывается с того места, где кончился прошлый прогон.
   */
  resumeAt: (at: number, play: boolean) => void
}

/**
 * Прогоняет сценарий до конца сразу, а дальше UI просто двигает курсор по истории.
 * Движок детерминирован, поэтому «шаг назад» и перемотка — это чтение готового снимка.
 */
export function useSimulation(scenario: MapScenario, seed: number, opts: { autoPause: boolean }): Playback {
  const sim = useMemo(() => new Simulation(scenario, seed).runToEnd(HARD_CAP), [scenario, seed])
  const last = sim.history.length - 1

  const [rawCursor, setCursorRaw] = useState(0)
  // Новый прогон может оказаться короче предыдущего, а сброс курсора произойдёт
  // только в эффекте — до него рендер обязан оставаться в пределах истории.
  const cursor = Math.min(rawCursor, last)
  const [playing, setPlaying] = useState(false)
  const [speed, setSpeed] = useState(4)
  const [pausedOn, setPausedOn] = useState<MapEvent | null>(null)

  const seen = useRef(new Set<MapEventType>())
  const resume = useRef<{ at: number; play: boolean } | null>(null)

  useEffect(() => {
    const r = resume.current
    resume.current = null
    setPausedOn(null)
    seen.current = new Set()
    if (!r) {
      setCursorRaw(0)
      setPlaying(false)
      return
    }
    const at = Math.min(r.at, last)
    if (r.play && last - at > 1) {
      setCursorRaw(at)
      setPlaying(true)
    } else {
      setCursorRaw(last)
      setPlaying(false)
    }
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

  const eventsUpTo = useMemo(() => sim.history.slice(1, cursor + 1).flatMap((s) => s.events), [sim, cursor])

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
    resumeAt: (at, play) => {
      resume.current = { at, play }
    },
  }
}
