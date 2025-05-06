export function startExam(examId: number) {
  const { $api } = useNuxtApp()

  return $api(`/api/exams/${examId}/start`, {
    method: 'PATCH',
  })
}
