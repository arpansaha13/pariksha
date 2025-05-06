export function endExam(examId: number) {
  const { $api } = useNuxtApp()

  return $api(`/api/exams/${examId}/end`, {
    method: 'PATCH',
  })
}
