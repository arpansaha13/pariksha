import type { ExamParticipantStatus } from '~/types/exam'

interface CheckParticipantAccessResponse {
  participant_status: ExamParticipantStatus
}

export async function checkParticipantAccess(examId: number) {
  const fetchOptions = getFetchOptions()
  fetchOptions.cache = 'no-cache'

  return $fetch<CheckParticipantAccessResponse>(
    `/api/exams/${examId}/participants/check`,
    fetchOptions
  )
}
