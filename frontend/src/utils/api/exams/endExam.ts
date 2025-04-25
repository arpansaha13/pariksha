export function endExam(examId: number) {
  return $fetch(`/api/exams/${examId}/end`, {
    method: 'PATCH',
    ...getFetchOptions(),
  })
}
