import { isNullOrUndefined } from '@arpansaha13/utils'
import { QuestionId, type Question } from '~/types'

export function useQuestion(questionId: ComputedRef<number | null>) {
  const fetchOptions = getFetchOptions()
  const { payload } = useNuxtApp()

  return useAsyncData(
    AsyncDataKeys.QUESTION(),
    async () => {
      if (isNullOrUndefined(questionId.value)) return Promise.resolve(null)
      if (questionId.value === QuestionId.ADD) return Promise.resolve(null)
      const data = await $fetch<Question>(
        `/api/questions/${questionId.value}`,
        fetchOptions
      )

      // A new key is not created in Nuxt Cache/Payload by itself when watched value changes
      payload.data[AsyncDataKeys.QUESTION(questionId.value)] = data
      return data
    },
    {
      watch: [questionId],
    }
  )
}
