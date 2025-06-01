import { isNullOrUndefined } from '@arpansaha13/utils'

export function useExamQuestion(questionId: ComputedRef<number | null>) {
  const { $api } = useNuxtApp()

  return useAsyncData(
    () => UseAsyncDataKeys.exam_question(questionId.value),
    async () => {
      if (isNullOrUndefined(questionId.value)) return Promise.resolve(null)

      const data = await $api<Question>(
        `/api/exams/questions/${questionId.value}`
      )

      // If there are no tags, backend returns null
      if (isNullOrUndefined(data.tags)) data.tags = []

      return data
    }
  )
}
