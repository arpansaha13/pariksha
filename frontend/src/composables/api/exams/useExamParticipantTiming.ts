interface ExamParticipantTiming {
  started_at: string
  scheduled_end_time: string
}

export function useExamParticipantTiming(examId: number) {
  const fetchOptions = getFetchOptions()

  return useAsyncData<ExamParticipantTiming>(
    AsyncDataKeys.EXAM_PARTICIPANT_TIMING(examId),
    () => $fetch(`/api/exams/${examId}/participants/timing`, fetchOptions)
  )
}
