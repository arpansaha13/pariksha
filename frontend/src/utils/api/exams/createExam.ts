import type { Exam, ExamAccessType } from '~/types/exam'

interface CreateExamBody {
  title: string
  starts_at: Date
  ends_at: Date
  type: ExamAccessType
  paper_id: number
  duration_minutes: number
}

export async function createExam(body: CreateExamBody) {
  const res = await $fetch<string>('/api/exams', {
    method: 'POST',
    body,
    ...getFetchOptions(),
  })

  return JSON.parse(res) as Exam
}
