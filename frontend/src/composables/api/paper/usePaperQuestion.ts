import { isNullOrUndefined } from '@arpansaha13/utils'

export function usePaperQuestion(
  questionId: QuestionId | ComputedRef<QuestionId | null>
) {
  const { $api } = useNuxtApp()

  return useAsyncData(
    () => UseAsyncDataKeys.paper_question(unref(questionId)),
    async () => {
      if (unref(questionId) === QUESTION_ID_ADD) return Promise.resolve(null)

      const data = await $api<Question>(`/api/questions/${unref(questionId)}`)

      // If there are no tags, backend returns null
      if (isNullOrUndefined(data.tags)) data.tags = []

      return data
    }
  )
}
