export function useExamParticipants(examId: ExamId) {
  const { $api } = useNuxtApp()

  return useAsyncData<ExamParticipantResponse[]>(
    AsyncDataKeys.EXAM_PARTICIPANTS(examId),
    () => $api(`/api/exams/${examId}/participants`)
  )
}
