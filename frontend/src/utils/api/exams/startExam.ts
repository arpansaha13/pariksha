export function startExam(examId: ExamId) {
  const { $api } = useNuxtApp()

  return $api(`/api/exams/${examId}/start`, {
    method: 'PATCH',
  })
}
