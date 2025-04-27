import { isNullOrUndefined } from '@arpansaha13/utils'
import type { AnswerMinimal } from '~/types'

export function useExamAnswer(examId: number, questionId: Ref<number | null>) {
  const fetchOptions = getFetchOptions()
  const { payload } = useNuxtApp()

  return useAsyncData(
    AsyncDataKeys.EXAM_ANSWER(),
    async () => {
      if (isNullOrUndefined(questionId.value)) {
        return Promise.resolve(null)
      }

      const data = await $fetch<AnswerMinimal>(
        `/api/exams/${examId}/questions/${questionId.value}/answer`,
        fetchOptions
      )
      // A new key is not created in Nuxt Cache/Payload by itself when watched value changes
      payload.data[AsyncDataKeys.EXAM_ANSWER(examId, questionId.value)] = data
      return data
    },
    {
      watch: [questionId],
    }
  )
}
