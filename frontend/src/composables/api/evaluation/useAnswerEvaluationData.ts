import { isNullOrUndefined } from '@arpansaha13/utils'
import type { EvaluationAnswer } from '~/types'

export function useAnswerEvaluationData(
  participantId: number,
  questionId: ComputedRef<number | null>
) {
  const { $api } = useNuxtApp()

  const useEvaluationAnswerKey = computed(() =>
    AsyncDataKeys.ANSWER_EVALUATION_DATA(participantId, questionId.value)
  )

  return useAsyncData(
    useEvaluationAnswerKey,
    () => {
      if (isNullOrUndefined(questionId.value)) {
        return Promise.resolve(null)
      }

      return $api<EvaluationAnswer>(
        `/api/participants/${participantId}/questions/${questionId.value}/evaluation-data`
      )
    },
    { server: false }
  )
}
