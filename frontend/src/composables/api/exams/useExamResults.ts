import type { ExamResult } from '~/types'

export function useExamResults(examId: number) {
  const { $api } = useNuxtApp()

  return useAsyncData<ExamResult[]>(AsyncDataKeys.EXAM_RESULTS(examId), () =>
    $api(`/api/exams/${examId}/results`)
  )
}
