import type { Exam } from '~/types/exam'

export function useExam(examId: number) {
  const fetchOptions = getFetchOptions()

  return useAsyncData<Exam>(AsyncDataKeys.EXAM(examId), () =>
    $fetch(`/api/exams/${examId}`, fetchOptions)
  )
}
