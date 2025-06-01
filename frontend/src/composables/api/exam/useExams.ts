export function useExams() {
  const { $api } = useNuxtApp()

  return useAsyncData<Exam[]>(UseAsyncDataKeys.exams, () => $api('/api/exams'))
}
