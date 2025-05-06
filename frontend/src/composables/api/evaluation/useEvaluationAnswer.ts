import { isNullOrUndefined } from '@arpansaha13/utils'
import type { Answer } from '~/types'

export function useEvaluationAnswer(
  participantId: number,
  questionId: ComputedRef<number | null>
) {
  const { $api } = useNuxtApp()

  const useEvaluationAnswerKey = computed(() =>
    AsyncDataKeys.EXAM_ANSWER(participantId, questionId.value)
  )

  return useAsyncData(
    useEvaluationAnswerKey,
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
