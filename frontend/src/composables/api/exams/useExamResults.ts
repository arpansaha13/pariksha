export function useExamResults(examId: number) {
  const { $api } = useNuxtApp()

  return useAsyncData(
    AsyncDataKeys.EXAM_RESULTS(examId),
    () => $api<ExamResult[]>(`/api/exams/${examId}/results`),
    {
      transform: res => {
        const byAnswerId: Record<ExamResult['id'], ExamResult> = {}

        for (const ans of res) {
          byAnswerId[ans.id] = ans
        }

        return byAnswerId
      },
    }
  )
}
