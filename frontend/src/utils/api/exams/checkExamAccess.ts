import type { ExamParticipantStatus, ExamPermission } from '~/types/exam'

interface CheckExamAccessResponse {
  access_type: ExamPermission
  participant_status: ExamParticipantStatus
}

export async function checkExamAccess(examId: number) {
  const fetchOptions = getFetchOptions()
  fetchOptions.cache = 'no-cache'

  return $fetch<CheckExamAccessResponse>(
    `/api/exams/${examId}/check`,
    fetchOptions
  )
}
