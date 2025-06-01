export function useExams() {
  const { $api } = useNuxtApp()

  return useAsyncData<Exam[]>(AsyncDataKeys.EXAMS, () => $api('/api/exams'))
}
