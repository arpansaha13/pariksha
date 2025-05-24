import { isNullOrUndefined } from '@arpansaha13/utils'

export function usePaperQuestion(
  questionId: QuestionId | ComputedRef<QuestionId | null>
) {
  const { $api } = useNuxtApp()

  const useQuestionKey = computed(() =>
    AsyncDataKeys.QUESTION(unref(questionId))
  )

  return useAsyncData(useQuestionKey, async () => {
    if (unref(questionId) === QUESTION_ID_ADD) return Promise.resolve(null)

    const data = await $api<Question>(`/api/questions/${unref(questionId)}`)

    // If there are no tags, backend returns null
    if (isNullOrUndefined(data.tags)) data.tags = []

    return data
  })
}
