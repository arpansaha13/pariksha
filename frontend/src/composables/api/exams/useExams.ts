import type { Exam } from '~/types/exam'

export function useExams() {
  const { $api } = useNuxtApp()

  return useAsyncData<Exam[]>(AsyncDataKeys.EXAMS, () => $api('/api/exams'))
}
