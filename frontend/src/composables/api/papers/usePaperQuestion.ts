import { isNullOrUndefined } from '@arpansaha13/utils'
import { QuestionId, type Question } from '~/types'

export function usePaperQuestion(questionId: ComputedRef<number | null>) {
  const fetchOptions = getFetchOptions()

  const useQuestionKey = computed(() =>
    AsyncDataKeys.QUESTION(questionId.value)
  )

  return useAsyncData(useQuestionKey, async () => {
    if (questionId.value === QuestionId.ADD) return Promise.resolve(null)

    const data = await $fetch<Question>(
      `/api/questions/${questionId.value}`,
      fetchOptions
    )

    // If there are no tags, backend returns null
    if (isNullOrUndefined(data.tags)) data.tags = []

    return data
  })
}
