import type { QuestionMinimal } from '~/types'

export function usePaperQuestions(paperId: number) {
  const fetchOptions = getFetchOptions()

  return useAsyncData(
    AsyncDataKeys.PAPERS_PAPER_QUESTIONS(paperId),
    () =>
      $fetch<QuestionMinimal[]>(
        `/api/papers/${paperId}/questions`,
        fetchOptions
      ),
    {
      transform: questions => {
        const byCategory = {} as Record<
          number | 'uncategorized',
          QuestionMinimal[]
        >

        for (const question of questions) {
          const categoryId = question.category_id ?? 'uncategorized'
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
