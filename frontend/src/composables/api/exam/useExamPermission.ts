export function useExamPermission(examId: ExamId) {
  const { $api } = useNuxtApp()

  return useAsyncData<ExamPermission>(
    UseAsyncDataKeys.exam_permission(examId),
    () => $api(`/api/exams/${examId}/permission`)
  )
}
