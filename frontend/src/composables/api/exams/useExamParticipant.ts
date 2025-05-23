export function useExamParticipant(examId: ExamId) {
  const { $api } = useNuxtApp()

  return useAsyncData<ExamParticipant>(
    AsyncDataKeys.EXAM_PARTICIPANT(examId),
    () => $api(`/api/exams/${examId}/participants/current`)
  )
}
