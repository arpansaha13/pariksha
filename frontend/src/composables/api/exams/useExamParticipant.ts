import type { ExamParticipant } from '~/types'

export function useExamParticipant(examId: number) {
  const { $api } = useNuxtApp()

  return useAsyncData<ExamParticipant>(
    AsyncDataKeys.EXAM_PARTICIPANT(examId),
    () => $api(`/api/exams/${examId}/participants/current`)
  )
}
