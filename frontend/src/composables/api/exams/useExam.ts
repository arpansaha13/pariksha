import type { Exam } from '~/types/exam'

export function useExam(examId: number) {
  const { $api } = useNuxtApp()

  return useAsyncData<Exam>(AsyncDataKeys.EXAM(examId), () =>
    $api(`/api/exams/${examId}`)
  )
}
