import { isNullOrUndefined } from '@arpansaha13/utils'
import { QuestionId, type Question } from '~/types'

export function useExamQuestion(questionId: ComputedRef<number | null>) {
  const fetchOptions = getFetchOptions()
  const { payload } = useNuxtApp()

  return useAsyncData(
    AsyncDataKeys.QUESTION(),
    async () => {
      if (isNullOrUndefined(questionId.value)) return Promise.resolve(null)
      if (questionId.value === QuestionId.ADD) return Promise.resolve(null)
      const data = await $fetch<Question>(
        `/api/exams/questions/${questionId.value}`,
        fetchOptions
      )

      // If there are no tags, backend returns null
      if (isNullOrUndefined(data.tags)) data.tags = []

      // A new key is not created in Nuxt Cache/Payload by itself when watched value changes
      payload.data[AsyncDataKeys.QUESTION(questionId.value)] = data
      return data
    },
    {
      watch: [questionId],
    }
  )
}
