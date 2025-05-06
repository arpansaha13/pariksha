import type { ExamCategory } from '~/types'

export function useExamCategories(examId: number) {
  const { $api } = useNuxtApp()

  return useAsyncData<ExamCategory[]>(
    AsyncDataKeys.EXAM_CATEGORIES(examId),
    () => $api<ExamCategory[]>(`/api/exams/${examId}/categories`),
    {
      transform: categories => categories.toSorted((a, b) => a.order - b.order),
    }
  )
}
