import { isNullOrUndefined } from '@arpansaha13/utils'

export function useExamQuestion(
  questionId: QuestionId | ComputedRef<QuestionId | null>
) {
  const { $api } = useNuxtApp()

  return useAsyncData(
    () => UseAsyncDataKeys.exam_question(unref(questionId)),
    async () => {
      if (isNullOrUndefined(unref(questionId))) return Promise.resolve(null)

      const data = await $api<Question>(
        `/api/exams/questions/${unref(questionId)}`
      )

      // If there are no tags, backend returns null
      if (isNullOrUndefined(data.tags)) data.tags = []

      return data
    }
  )
}
