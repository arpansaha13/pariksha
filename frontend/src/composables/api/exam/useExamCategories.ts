export function useExamCategories(examId: ExamId) {
  const { $api } = useNuxtApp()

  return useAsyncData<ExamCategory[]>(
    UseAsyncDataKeys.exam_categories(examId),
    () => $api<ExamCategory[]>(`/api/exams/${examId}/categories`),
    {
      transform: categories => categories.toSorted((a, b) => a.order - b.order),
    }
  )
}
