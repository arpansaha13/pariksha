import type { Exam } from '~/types/exam'

interface UpdateExamBody {
  title?: string
  starts_at?: Date
  ends_at?: Date
  type?: string
  duration_minutes?: number
}

export async function updateExam(examId: number, body: UpdateExamBody) {
  const { $api } = useNuxtApp()

  const res = await $api<string>(`/api/exams/${examId}`, {
    method: 'PATCH',
    body,
  })

  refreshNuxtData(AsyncDataKeys.EXAM(examId))
  return JSON.parse(res) as Exam
}
