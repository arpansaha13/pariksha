import { isNullOrUndefined } from '@arpansaha13/utils'

export function usePaperQuestion(questionId: ComputedRef<number | null>) {
  const { $api } = useNuxtApp()

  const useQuestionKey = computed(() =>
    AsyncDataKeys.QUESTION(questionId.value)
  )

  return useAsyncData(useQuestionKey, async () => {
    if (questionId.value === QUESTION_ID_ADD) return Promise.resolve(null)

    const data = await $api<Question>(`/api/questions/${questionId.value}`)

    // If there are no tags, backend returns null
    if (isNullOrUndefined(data.tags)) data.tags = []

    return data
  })
}
