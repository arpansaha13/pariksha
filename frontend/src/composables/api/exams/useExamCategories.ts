import type { ExamCategory } from '~/types'

export function useExamCategories(examId: number) {
  const fetchOptions = getFetchOptions()

  return useAsyncData<ExamCategory[]>(
    AsyncDataKeys.EXAM_CATEGORIES(examId),
    () =>
      $fetch<ExamCategory[]>(`/api/exams/${examId}/categories`, fetchOptions),
    {
      transform: categories => categories.toSorted((a, b) => a.order - b.order),
    }
  )
}
