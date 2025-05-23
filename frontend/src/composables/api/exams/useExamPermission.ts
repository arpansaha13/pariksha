export function useExamPermission(examId: number) {
  const { $api } = useNuxtApp()

  return useAsyncData<ExamPermission>(
    AsyncDataKeys.EXAM_PERMISSION(examId),
    () => $api(`/api/exams/${examId}/permission`)
  )
}
