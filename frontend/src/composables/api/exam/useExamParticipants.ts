export function useExamParticipants(examId: ExamId) {
  const { $api } = useNuxtApp()

  return useAsyncData<ExamParticipantResponse[]>(
    UseAsyncDataKeys.exam_participants(examId),
    () => $api(`/api/exams/${examId}/participants`)
  )
}
