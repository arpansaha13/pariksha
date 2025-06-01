interface ExamParticipantById {
  readonly id: ExamParticipantId
  readonly status: ExamParticipantStatus
}

export function useExamParticipantById(participantId: ExamParticipantId) {
  const { $api } = useNuxtApp()

  return useAsyncData<ExamParticipantById>(
    AsyncDataKeys.EXAM_PARTICIPANT_BY_ID(participantId),
    () => $api(`/api/participants/${participantId}`)
  )
}
