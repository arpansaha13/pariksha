import type { Question } from '~/types'

export function usePaperQuestions(paperId: number) {
  const fetchOptions = getFetchOptions()

  return useAsyncData(
    AsyncDataKeys.PAPERS_PAPER_QUESTIONS(paperId),
    () => $fetch<Question[]>(`/api/papers/${paperId}/questions`, fetchOptions),
    {
      transform: questions => {
        const byCategory = {} as Record<number | 'uncategorized', Question[]>

        for (const question of questions) {
          const categoryId = question.category?.id ?? 'uncategorized'
          if (!byCategory[categoryId]) {
            byCategory[categoryId] = []
          }
          byCategory[categoryId].push(question)
        }

        return byCategory
      },
    }
  )

  // Not working in server-side
  // transform: questions => {
  //   return Object.groupBy(questions, q => q.category?.id ?? 'uncategorized')
  // },
}
