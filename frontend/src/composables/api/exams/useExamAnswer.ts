import { isNullOrUndefined } from '@arpansaha13/utils'
import type { AnswerMinimal } from '~/types'

export function useExamAnswer(
  examId: number,
  questionId: ComputedRef<number | null>
) {
  const { $api } = useNuxtApp()

  const useExamAnswerKey = computed(() =>
    AsyncDataKeys.EXAM_ANSWER(examId, questionId.value)
  )

  return useAsyncData(
    useExamAnswerKey,
    () => {
      if (isNullOrUndefined(questionId.value)) {
        return Promise.resolve(null)
      }

      return $api<AnswerMinimal>(
        `/api/exams/${examId}/questions/${questionId.value}/answer`
      )
    },
    { server: false }
  )
}
