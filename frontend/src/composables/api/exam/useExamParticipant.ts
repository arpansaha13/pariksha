export function useExamParticipant(examId: ExamId) {
  const { $api } = useNuxtApp()

  return useAsyncData<ExamParticipant>(
    UseAsyncDataKeys.exam_participant(examId),
    () => $api(`/api/exams/${examId}/participants/current`)
  )
}
