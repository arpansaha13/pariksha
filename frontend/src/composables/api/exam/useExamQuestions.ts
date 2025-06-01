export function useExamQuestions(examId: ExamId) {
  const { $api } = useNuxtApp()

  return useAsyncData(
    UseAsyncDataKeys.exam_questions(examId),
    () => $api<ExamQuestionMinimal[]>(`/api/exams/${examId}/questions`),
    {
      transform: questions => {
        const byCategory = {} as Record<CategoryId, ExamQuestionMinimal[]>

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
