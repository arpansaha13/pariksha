export function startExam(examId: number) {
  return $fetch(`/api/exams/${examId}/start`, {
    method: 'PATCH',
    ...getFetchOptions(),
  })
}
