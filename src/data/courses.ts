/**
 * Курсы и их темы. Готовая тема — это лекция в src/content/lectures/<course>/<slug>.mdx;
 * здесь же лежат анонсы тем, которых ещё нет, чтобы на главной был виден план.
 */
export interface Course {
  id: string
  title: string
  description: string
  planned: { title: string; blurb: string }[]
}

export const COURSES: Course[] = [
  {
    id: 'go-runtime',
    title: 'Рантайм Go',
    description:
      'Что происходит под программой на Go: как горутины попадают на процессор, кто собирает мусор, где живут переменные и как устроены каналы.',
    planned: [
      { title: 'Сборщик мусора', blurb: 'Трёхцветная разметка, барьер записи, GOGC и GOMEMLIMIT, паузы STW.' },
      { title: 'Память', blurb: 'Стек против кучи, escape analysis, mcache/mcentral/mheap, классы размеров.' },
      { title: 'Каналы', blurb: 'hchan изнутри: буфер, sendq и recvq, select, закрытие.' },
    ],
  },
]
