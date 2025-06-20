import { isNullOrUndefined } from '@arpansaha13/utils'
import { defineStore } from 'pinia'

interface ExamStore {
  examId: ExamId
  areAnswerStatesPrepared: boolean
  answerFetched: Record<QuestionId, boolean>
  mcqAnswerStates: Record<QuestionId, MCQAnswer>
  subjectiveAnswerStates: Record<QuestionId, SubjectiveAnswer>
  codingAnswerStates: Record<QuestionId, CodingAnswer>
  savedMcqAnswerStates: Record<QuestionId, MCQAnswer>
  savedSubjectiveAnswerStates: Record<QuestionId, SubjectiveAnswer>
  savedCodingAnswerStates: Record<QuestionId, CodingAnswer>
}

export const useExamStore = defineStore(examStoreId, {
  state: (): ExamStore => ({
    examId: '' as ExamId,
    areAnswerStatesPrepared: false,
    answerFetched: {},
    mcqAnswerStates: {},
    subjectiveAnswerStates: {},
    codingAnswerStates: {},
    savedMcqAnswerStates: {},
    savedSubjectiveAnswerStates: {},
    savedCodingAnswerStates: {},
  }),
  actions: {
    prepare(groupedQuestions: Record<CategoryId, ExamQuestionMinimal[]>) {
      if (this.areAnswerStatesPrepared) return
      this.areAnswerStatesPrepared = true

      for (const questionMinimals of Object.values(groupedQuestions)) {
        for (const questionMinimal of questionMinimals) {
          const qid = questionMinimal.id
          if (questionMinimal.type === QuestionType.MCQ) {
            this.mcqAnswerStates[qid] = {
              optionIndex: undefined,
            }
            this.savedMcqAnswerStates[qid] = {
              ...this.mcqAnswerStates[qid],
            }
          } else if (questionMinimal.type === QuestionType.SUBJECTIVE) {
            this.subjectiveAnswerStates[qid] = {
              text: '',
            }
            this.savedSubjectiveAnswerStates[qid] = {
              ...this.subjectiveAnswerStates[qid],
            }
          } else {
            if (!isNullOrUndefined(this.codingAnswerStates[qid])) return

            this.codingAnswerStates[qid] = {
              code: '',
            }
            this.savedCodingAnswerStates[qid] = {
              ...this.codingAnswerStates[qid],
            }
          }
        }
      }
    },
    clearMcqSelection(qid: QuestionId) {
      if (
        isNullOrUndefined(qid) ||
        isNullOrUndefined(this.mcqAnswerStates[qid])
      )
        return
      this.mcqAnswerStates[qid].optionIndex = undefined
    },

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

    saveCodingAnswer(
      questionId: QuestionId,
      savedAnswer: CodingAnswer,
      newAnswer: CodingAnswer
    ) {
      const upsertAnswerBody = {
        question_id: questionId,
        answer: null as CodingAnswer | null,
      }

      const currentText = savedAnswer ? savedAnswer.code : ''
      const answerCode = newAnswer?.code ?? ''

      if (savedAnswer && !answerCode) {
        return upsertAnswer(this.examId, upsertAnswerBody)
      }

      if (!currentText && !answerCode) {
        return Promise.resolve(null)
      }

      if (currentText === answerCode) {
        return Promise.resolve(null)
      }

      upsertAnswerBody.answer = { code: answerCode }
      return upsertAnswer(this.examId, upsertAnswerBody)
    },
    saveUpdatedAnswers() {
      const promises = []

      for (const [qid, mcqAnswer] of Object.entries(this.mcqAnswerStates)) {
        const savedState = this.savedMcqAnswerStates[qid as QuestionId]

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
      for (const [qid, subjectiveAnswer] of Object.entries(
        this.subjectiveAnswerStates
      )) {
        const savedState = this.savedSubjectiveAnswerStates[qid as QuestionId]

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
      for (const [qid, codingAnswer] of Object.entries(
        this.codingAnswerStates
      )) {
        const savedState = this.savedCodingAnswerStates[qid as QuestionId]

        if (savedState.code !== codingAnswer.code) {
          promises.push(
            this.saveCodingAnswer(
              qid as QuestionId,
              savedState,
              codingAnswer
            ).then(res => {
              if (isNullOrUndefined(res) || isNullOrUndefined(res.answer)) {
                savedState.code = ''
              } else {
                savedState.code = (res.answer as CodingAnswer).code
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
      } else if (qType === QuestionType.CODING) {
        this.codingAnswerStates[qid] = {
          code: (answerContent as CodingAnswer).code,
        }
        this.savedCodingAnswerStates[qid] = {
          ...this.codingAnswerStates[qid],
        }
      }
    },
  },
})
