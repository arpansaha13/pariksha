import type { ExamQuestionMinimal } from '~/types'

export function useExamQuestions(examId: number) {
  const fetchOptions = getFetchOptions()

  return useAsyncData(
    AsyncDataKeys.EXAM_QUESTIONS(examId),
    () =>
      $fetch<ExamQuestionMinimal[]>(
        `/api/exams/${examId}/questions`,
        fetchOptions
      ),
    {
      transform: questions => {
        const byCategory = {} as Record<number, ExamQuestionMinimal[]>

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
}
