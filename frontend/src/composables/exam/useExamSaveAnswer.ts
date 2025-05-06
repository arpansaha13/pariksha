import { isNullOrUndefined } from '@arpansaha13/utils'
import {
  QuestionType,
  type AnswerMinimal,
  type ExamQuestionMinimal,
  type GeneralAnswer,
  type MCQAnswer,
  type Question,
} from '~/types'

interface UseExamSaveAnswerArgs {
  examId: number
  answer: Ref<AnswerMinimal | null>
  groupedQuestions: Ref<Record<number, ExamQuestionMinimal[]> | null>
}

type MergedAnswer = MCQAnswer & GeneralAnswer

export function useExamSaveAnswer(args: UseExamSaveAnswerArgs) {
  const { examId, groupedQuestions, answer } = args

  const answerStates = reactive<Record<number, MergedAnswer>>({})

  for (const questionMinimals of Object.values(groupedQuestions.value!)) {
    for (const questionMinimal of questionMinimals) {
      answerStates[questionMinimal.id] = {
        optionIndex: undefined,
        text: '',
      }
    }
  }

  function updateAnswerState(newAnswer: AnswerMinimal | null) {
    if (isNullOrUndefined(newAnswer)) return
    if (isNullOrUndefined(newAnswer.answer)) {
      answerStates[newAnswer.question_id].optionIndex = undefined
      answerStates[newAnswer.question_id].text = ''
      return
    }

    answerStates[newAnswer.question_id].optionIndex = (
      newAnswer.answer as MCQAnswer
    ).optionIndex

    answerStates[newAnswer.question_id].text =
      (newAnswer.answer as GeneralAnswer).text ?? ''
  }

  watchImmediate(answer, newAnswer => {
    updateAnswerState(newAnswer)
  })

  /** Save answer for a specific question */
  function saveAnswer(questionToSave: Question) {
    const answerState = answerStates[questionToSave.id]
    const { data: currentAnswer } = useNuxtData<AnswerMinimal | null>(
      AsyncDataKeys.EXAM_ANSWER(examId, questionToSave.id)
    )

    const upsertAnswerBody = {
      question_id: questionToSave.id,
      answer: null as MCQAnswer | GeneralAnswer | null,
    }

    if (questionToSave.type === QuestionType.MCQ) {
      const currentOptionIndex = currentAnswer.value?.answer
        ? (currentAnswer.value.answer as MCQAnswer).optionIndex
        : undefined

      // If there's a saved answer but current state is empty, clear the answer
      if (
        currentAnswer.value?.answer &&
        isNullOrUndefined(answerState.optionIndex)
      ) {
        return upsertAnswer(examId, upsertAnswerBody)
      }

      // Both empty, no action needed
      if (
        isNullOrUndefined(answerState.optionIndex) &&
        isNullOrUndefined(currentOptionIndex)
      ) {
        return
      }

      // Both have same value, no action needed
      if (answerState.optionIndex === currentOptionIndex) {
        return
      }

      // One is empty and other has value, or they have different values
      upsertAnswerBody.answer = { optionIndex: answerState.optionIndex }
    } else {
      const currentText = currentAnswer.value?.answer
        ? (currentAnswer.value.answer as GeneralAnswer).text
        : ''
      const answerText = answerState.text ?? ''

      // If there's a saved answer but current state is empty, clear the answer
      if (currentAnswer.value?.answer && !answerText) {
        return upsertAnswer(examId, upsertAnswerBody)
      }
      // Both empty, no action needed
      if (!currentText && !answerText) {
        return
      }

      // Both have same value, no action needed
      if (currentText === answerText) {
        return
      }

      // One is empty and other has value, or they have different values
      upsertAnswerBody.answer = { text: answerText }
    }

    return upsertAnswer(examId, upsertAnswerBody)
  }

  return { answerStates, saveAnswer }
}
