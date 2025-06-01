export function useExam(examId: ExamId) {
  const { $api } = useNuxtApp()

  return useAsyncData<Exam>(AsyncDataKeys.EXAM(examId), () =>
    $api(`/api/exams/${examId}`)
  )
}
