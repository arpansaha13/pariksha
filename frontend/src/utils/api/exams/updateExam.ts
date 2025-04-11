import type { Exam } from '~/types/exam'

interface UpdateExamBody {
  title?: string
  starts_at?: Date
  ends_at?: Date
  type?: string
}

export async function updateExam(examId: number, body: UpdateExamBody) {
  const res = await $fetch<string>(`/api/exams/${examId}`, {
    method: 'PATCH',
    body,
    ...getFetchOptions(),
  })

  return JSON.parse(res) as Exam
}
