interface ExamParticipantById {
  readonly id: ExamParticipantId
  readonly status: ExamParticipantStatus
}

export function useExamParticipantById(participantId: ExamParticipantId) {
  const { $api } = useNuxtApp()

  return useAsyncData<ExamParticipantById>(
    UseAsyncDataKeys.participant_by_id(participantId),
    () => $api(`/api/participants/${participantId}`)
  )
}
