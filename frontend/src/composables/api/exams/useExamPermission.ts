import type { ExamPermission } from '~/types/exam'

export function useExamPermission(examId: number) {
  const fetchOptions = getFetchOptions()
  fetchOptions.cache = 'no-cache'

  return useAsyncData<ExamPermission>(AsyncDataKeys.EXAM_ACCESS(examId), () =>
    $fetch(`/api/exams/${examId}/permission`, fetchOptions)
  )
}
