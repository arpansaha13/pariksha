import type { ExamPermission } from '~/types/exam'

interface CheckExamAccessResponse {
  access_type: ExamPermission
}

export async function checkExamAccess(examId: number) {
  const fetchOptions = getFetchOptions()
  fetchOptions.cache = 'no-cache'

  return $fetch<CheckExamAccessResponse>(
    `/api/exams/${examId}/check`,
    fetchOptions
  )
}
