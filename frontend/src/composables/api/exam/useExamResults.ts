export function useExamResults(examId: ExamId) {
  const { $api } = useNuxtApp()

  return useAsyncData(
    UseAsyncDataKeys.exam_results(examId),
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
