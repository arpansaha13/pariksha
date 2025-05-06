import { isNullOrUndefined } from '@arpansaha13/utils'
import type { AnswerMinimal } from '~/types'

export function useExamAnswer(
  examId: number,
  questionId: ComputedRef<number | null>
) {
  const fetchOptions = getFetchOptions()

  const useExamAnswerKey = computed(() =>
    AsyncDataKeys.EXAM_ANSWER(examId, questionId.value)
  )

  return useAsyncData(
    useExamAnswerKey,
    () => {
      if (isNullOrUndefined(questionId.value)) {
        return Promise.resolve(null)
      }

      return $fetch<AnswerMinimal>(
        `/api/exams/${examId}/questions/${questionId.value}/answer`,
        fetchOptions
      )
    },
    { server: false }
  )
}
