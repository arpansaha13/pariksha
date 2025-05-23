export function endExam(examId: ExamId) {
  const { $api } = useNuxtApp()

  return $api(`/api/exams/${examId}/end`, {
    method: 'PATCH',
  })
}
