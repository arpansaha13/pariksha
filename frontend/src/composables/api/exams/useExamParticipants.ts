import type { ExamParticipantResponse } from '~/types'

export function useExamParticipants(examId: number) {
  const { $api } = useNuxtApp()

  return useAsyncData<ExamParticipantResponse[]>(
    AsyncDataKeys.EXAM_PARTICIPANTS(examId),
    () => $api(`/api/exams/${examId}/participants`)
  )
}
