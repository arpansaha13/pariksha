import { isNullOrUndefined } from '@arpansaha13/utils'
import type { Answer } from '~/types'

export function useEvaluationAnswer(
  participantId: number,
  questionId: ComputedRef<number | null>
) {
  const fetchOptions = getFetchOptions()

  const useEvaluationAnswerKey = computed(() =>
    AsyncDataKeys.EXAM_ANSWER(participantId, questionId.value)
  )

  return useAsyncData(
    useEvaluationAnswerKey,
    () => {
      if (isNullOrUndefined(questionId.value)) {
        return Promise.resolve(null)
      }

      return $fetch<Answer>(
        `/api/participants/${participantId}/questions/${questionId.value}/answer`,
        fetchOptions
      )
    },
    { server: false }
  )
}
