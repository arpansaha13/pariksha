import type { ExamParticipantById } from '~/types'

export function useExamParticipantById(participantId: number) {
  const { $api } = useNuxtApp()

  return useAsyncData<ExamParticipantById>(
    AsyncDataKeys.EXAM_PARTICIPANT(participantId),
    () => $api(`/api/participants/${participantId}`)
  )
}
