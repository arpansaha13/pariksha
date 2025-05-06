import type { ExamQuestionMinimal } from '~/types'

export function useExamQuestions(examId: number) {
  const { $api } = useNuxtApp()

  return useAsyncData(
    AsyncDataKeys.EXAM_QUESTIONS(examId),
    () => $api<ExamQuestionMinimal[]>(`/api/exams/${examId}/questions`),
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
