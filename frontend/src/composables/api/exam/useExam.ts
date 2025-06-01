export function useExam(examId: ExamId) {
  const { $api } = useNuxtApp()

  return useAsyncData<Exam>(UseAsyncDataKeys.exam(examId), () =>
    $api(`/api/exams/${examId}`)
  )
}
