import type { Exam } from '~/types/exam'

export function useExams() {
  const fetchOptions = getFetchOptions()

  return useAsyncData<Exam[]>(AsyncDataKeys.EXAMS, () =>
    $fetch('/api/exams', fetchOptions)
  )
}
