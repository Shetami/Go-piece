/** id терминов, упомянутых в исходнике MDX через <Term id="...">, в порядке первого упоминания. */
export function termIdsIn(body: string | undefined): string[] {
  if (!body) return []
  const ids = new Set<string>()
  for (const m of body.matchAll(/<Term\s+id=["']([^"']+)["']/g)) {
    if (m[1]) ids.add(m[1])
  }
  return [...ids]
}
