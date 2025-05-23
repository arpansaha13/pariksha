export function usePaperQuestions(paperId: number) {
  const { $api } = useNuxtApp()

  return useAsyncData(
    AsyncDataKeys.PAPERS_PAPER_QUESTIONS(paperId),
    () => $api<QuestionMinimal[]>(`/api/papers/${paperId}/questions`),
    {
      transform: questions => {
        const byCategory = {} as Record<number, QuestionMinimal[]>

        for (const question of questions) {
          const categoryId = question.category_id
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
  //   return Object.groupBy(questions, q => q.category.id)
  // },
}
