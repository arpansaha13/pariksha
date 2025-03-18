import { isNullOrUndefined } from '@arpansaha13/utils'
import { QuestionId, type Question } from '~/types'

export function useQuestion(questionId: ComputedRef<number | null>) {
  const fetchOptions = getFetchOptions()

  return useAsyncData(
    AsyncDataKeys.QUESTION(questionId.value),
    () => {
      if (isNullOrUndefined(questionId.value)) return Promise.resolve(null)
      if (questionId.value === QuestionId.ADD) return Promise.resolve(null)
      return $fetch<Question>(
        `/api/questions/${questionId.value}`,
        fetchOptions
      )
    },
    {
      watch: [questionId],
    }
  )
}
