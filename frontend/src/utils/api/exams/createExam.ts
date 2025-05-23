interface CreateExamBody {
  title: string
  starts_at: Date
  ends_at: Date
  type: ExamAccessType
  paper_id: PaperId
  duration_minutes: number
}

export async function createExam(body: CreateExamBody) {
  const { $api } = useNuxtApp()

  const res = await $api<string>('/api/exams', {
    method: 'POST',
    body,
  })

  return JSON.parse(res) as Exam
}
