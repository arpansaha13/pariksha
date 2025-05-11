import { isNullOrUndefined } from '@arpansaha13/utils'
import type { Answer } from '~/types'

export function useAnswerForEvaluation(
  participantId: number,
  questionId: ComputedRef<number | null>
) {
  const { $api } = useNuxtApp()

  const useAnswerForEvaluationKey = computed(() =>
    AsyncDataKeys.EVALUATION_ANSWER(participantId, questionId.value)
  )

  return useAsyncData(
    useAnswerForEvaluationKey,
    () => {
      if (isNullOrUndefined(questionId.value)) {
        return Promise.resolve(null)
      }

      return $api<Answer>(
        `/api/participants/${participantId}/questions/${questionId.value}/answer`
      )
    },
    { server: false }
  )
}
