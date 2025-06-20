import { isNullOrUndefined } from '@arpansaha13/utils'
import { defineStore } from 'pinia'

interface ExamStore {
  examId: ExamId
  answerFetched: Record<string, boolean>
  mcqAnswerStates: Record<QuestionId, MCQAnswer | undefined>
  subjectiveAnswerStates: Record<QuestionId, SubjectiveAnswer | undefined>
  savedMcqAnswerStates: Record<string, MCQAnswer>
  savedSubjectiveAnswerStates: Record<string, SubjectiveAnswer>
}

export const useExamStore = defineStore(examStoreId, {
  state: (): ExamStore => ({
    examId: '' as ExamId,
    answerFetched: {},
    mcqAnswerStates: {},
    subjectiveAnswerStates: {},
    savedMcqAnswerStates: {},
    savedSubjectiveAnswerStates: {},
  }),
  actions: {
    clearMcqSelection(qid: QuestionId) {
      if (
        isNullOrUndefined(qid) ||
        isNullOrUndefined(this.mcqAnswerStates[qid])
      )
        return
      this.mcqAnswerStates[qid].optionIndex = undefined
    },

    /** Save answer for a MCQ question */
    saveMcqAnswer(
      questionId: QuestionId,
      savedAnswer: MCQAnswer,
      newAnswer: MCQAnswer
    ) {
      const upsertAnswerBody = {
        question_id: questionId,
        answer: null as MCQAnswer | null,
      }

      const currentOptionIndex = savedAnswer
        ? savedAnswer.optionIndex
        : undefined

      if (savedAnswer && isNullOrUndefined(newAnswer.optionIndex)) {
        return upsertAnswer(this.examId, upsertAnswerBody)
      }

      if (
        isNullOrUndefined(newAnswer.optionIndex) &&
        isNullOrUndefined(currentOptionIndex)
      ) {
        return Promise.resolve(null)
      }

      if (newAnswer.optionIndex === currentOptionIndex) {
        return Promise.resolve(null)
      }

      upsertAnswerBody.answer = { optionIndex: newAnswer.optionIndex }
      return upsertAnswer(this.examId, upsertAnswerBody)
    },

    /** Save answer for a subjective question */
    saveSubjectiveAnswer(
      questionId: QuestionId,
      savedAnswer: SubjectiveAnswer,
      newAnswer: SubjectiveAnswer
    ) {
      const upsertAnswerBody = {
        question_id: questionId,
        answer: null as SubjectiveAnswer | null,
      }

      const currentText = savedAnswer ? savedAnswer.text : ''
      const answerText = newAnswer?.text ?? ''

      if (savedAnswer && !answerText) {
        return upsertAnswer(this.examId, upsertAnswerBody)
      }

      if (!currentText && !answerText) {
        return Promise.resolve(null)
      }

      if (currentText === answerText) {
        return Promise.resolve(null)
      }

      upsertAnswerBody.answer = { text: answerText }
      return upsertAnswer(this.examId, upsertAnswerBody)
    },
    saveUpdatedAnswers() {
      const promises = []

      for (const entry of Object.entries(this.mcqAnswerStates)) {
        const qid = entry[0]
        const mcqAnswer = entry[1]!
        const savedState = this.savedMcqAnswerStates[qid]!

        if (savedState.optionIndex !== mcqAnswer.optionIndex) {
          promises.push(
            this.saveMcqAnswer(qid as QuestionId, savedState, mcqAnswer).then(
              res => {
                if (isNullOrUndefined(res) || isNullOrUndefined(res.answer)) {
                  savedState.optionIndex = undefined
                } else {
                  savedState.optionIndex = (res.answer as MCQAnswer).optionIndex
                }
              }
            )
          )
        }
      }
      for (const entry of Object.entries(this.subjectiveAnswerStates)) {
        const qid = entry[0]
        const subjectiveAnswer = entry[1]!
        const savedState = this.savedSubjectiveAnswerStates[qid]!

        if (savedState.text !== subjectiveAnswer.text) {
          promises.push(
            this.saveSubjectiveAnswer(
              qid as QuestionId,
              savedState,
              subjectiveAnswer
            ).then(res => {
              if (isNullOrUndefined(res) || isNullOrUndefined(res.answer)) {
                savedState.text = ''
              } else {
                savedState.text = (res.answer as SubjectiveAnswer).text
              }
            })
          )
        }
      }

      return promises
    },

    setAnswer(
      qid: QuestionId,
      qType: QuestionType,
      answerContent: AnswerMinimal['answer']
    ) {
      if (qType === QuestionType.MCQ) {
        this.mcqAnswerStates[qid] = {
          optionIndex: (answerContent as MCQAnswer).optionIndex,
        }
        this.savedMcqAnswerStates[qid] = { ...this.mcqAnswerStates[qid] }
      } else if (qType === QuestionType.SUBJECTIVE) {
        this.subjectiveAnswerStates[qid] = {
          text: (answerContent as SubjectiveAnswer).text,
        }
        this.savedSubjectiveAnswerStates[qid] = {
          ...this.subjectiveAnswerStates[qid],
        }
      }
    },
  },
})
