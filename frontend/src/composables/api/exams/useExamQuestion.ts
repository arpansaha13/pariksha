import { isNullOrUndefined } from '@arpansaha13/utils'
import type { Question } from '~/types'

export function useExamQuestion(questionId: ComputedRef<number | null>) {
  const fetchOptions = getFetchOptions()

  const useExamQuestionKey = computed(() =>
    AsyncDataKeys.EXAM_QUESTION(questionId.value)
  )

  return useAsyncData(useExamQuestionKey, async () => {
    if (isNullOrUndefined(questionId.value)) return Promise.resolve(null)

    const data = await $fetch<Question>(
      `/api/exams/questions/${questionId.value}`,
      fetchOptions
    )

    // If there are no tags, backend returns null
    if (isNullOrUndefined(data.tags)) data.tags = []

    return data
  })
}
