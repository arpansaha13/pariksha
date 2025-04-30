import type { ExamParticipantResponse } from '~/types'

export function useExamParticipants(examId: number) {
  const fetchOptions = getFetchOptions()

  return useAsyncData<ExamParticipantResponse[]>(
    AsyncDataKeys.EXAM_PARTICIPANTS(examId),
    () => $fetch(`/api/exams/${examId}/participants`, fetchOptions)
  )
}
