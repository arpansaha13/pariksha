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
  question: Ref<Question | null>
  answer: Ref<AnswerMinimal | null>
  groupedQuestions: Ref<Record<number, ExamQuestionMinimal[]> | null>
}

type MergedAnswer = MCQAnswer & GeneralAnswer

export function useExamSaveAnswer(args: UseExamSaveAnswerArgs) {
  const { examId, question, groupedQuestions, answer } = args

  const { payload } = useNuxtApp()
  const answerStates = reactive<Record<number, MergedAnswer>>({})

  for (const questionMinimals of Object.values(groupedQuestions.value!)) {
    for (const questionMinimal of questionMinimals) {
      answerStates[questionMinimal.id] = {
        optionIndex: undefined,
        text: '',
      }
    }
  }

  watchImmediate(answer, newAnswer => {
    if (isNullOrUndefined(newAnswer)) return

    answerStates[newAnswer.question_id].optionIndex = (
      newAnswer.answer as MCQAnswer
    ).optionIndex

    answerStates[newAnswer.question_id].text =
      (newAnswer.answer as GeneralAnswer).text ?? ''
  })

  // Save answer for a specific question
  function saveAnswer(questionToSave: Question) {
    const answerState = answerStates[questionToSave.id]
    const currentAnswer = payload.data[
      AsyncDataKeys.EXAM_ANSWER(examId, questionToSave.id)
    ] as AnswerMinimal | null

    const upsertAnswerBody = {} as MergedAnswer

    if (questionToSave.type === QuestionType.MCQ) {
      const isEmpty = isNullOrUndefined(answerState.optionIndex)
      const isUnchanged =
        !isNullOrUndefined(currentAnswer) &&
        answerState.optionIndex ===
          (currentAnswer.answer as MCQAnswer).optionIndex

      if (isEmpty || isUnchanged) return

      upsertAnswerBody.optionIndex = answerState.optionIndex
    } else {
      const isEmpty = !answerState.text
      const isUnchanged =
        !isNullOrUndefined(currentAnswer) &&
        answerState.text === (currentAnswer.answer as GeneralAnswer).text

      if (isEmpty || isUnchanged) return

      upsertAnswerBody.text = answerState.text
    }

    return upsertAnswer(examId, {
      question_id: questionToSave.id,
      answer: upsertAnswerBody,
    })
  }

  watch(question, (_, oldQuestion) => {
    if (isNullOrUndefined(oldQuestion)) return
    return saveAnswer(oldQuestion)

    // NOTE: The last question may not be saved if user doesn't navigate
    // to any other question before submitting.
    // So manually call `saveAnswer` before submission.
  })

  return { answerStates, saveAnswer }
}
