export function useExamPermission(examId: ExamId) {
  const { $api } = useNuxtApp()

  return useAsyncData<ExamPermission>(
    AsyncDataKeys.EXAM_PERMISSION(examId),
    () => $api(`/api/exams/${examId}/permission`)
  )
}
