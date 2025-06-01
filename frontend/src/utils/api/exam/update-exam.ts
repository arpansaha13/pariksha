interface UpdateExamBody {
  title?: string
  starts_at?: Date
  ends_at?: Date
  type?: string
  duration_minutes?: number
}

export async function updateExam(examId: ExamId, body: UpdateExamBody) {
  const { $api } = useNuxtApp()

  const res = await $api<string>(`/api/exams/${examId}`, {
    method: 'PATCH',
    body,
  })

  refreshNuxtData(UseAsyncDataKeys.exam(examId))
  return JSON.parse(res) as Exam
}
