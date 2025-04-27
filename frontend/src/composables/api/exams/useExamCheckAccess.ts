import type { ExamParticipantStatus, ExamPermission } from '~/types/exam'

interface CheckExamAccessResponse {
  access_type: ExamPermission
  participant_status: ExamParticipantStatus
}

export function useExamCheckAccess(examId: number) {
  const fetchOptions = getFetchOptions()
  fetchOptions.cache = 'no-cache'

  return useAsyncData<CheckExamAccessResponse>(
    AsyncDataKeys.EXAM_ACCESS(examId),
    () => $fetch(`/api/exams/${examId}/check`, fetchOptions)
  )
}
