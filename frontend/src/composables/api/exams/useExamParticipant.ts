import type { ExamParticipant } from '~/types'

export function useExamParticipant(examId: number) {
  const fetchOptions = getFetchOptions()

  return useAsyncData<ExamParticipant>(
    AsyncDataKeys.EXAM_PARTICIPANT(examId),
    () => $fetch(`/api/exams/${examId}/participants/current`, fetchOptions)
  )
}
