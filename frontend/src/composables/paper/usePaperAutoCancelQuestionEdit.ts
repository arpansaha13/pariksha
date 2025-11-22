import { isEqual } from 'lodash-es'

interface UsePaperAutoCancelQuestionEditArgs {
  question: Ref<Question | null>
  editQuestionFormStates: Record<QuestionId, MergedQuestion | null>
  decUnsavedCount: (categoryId: CategoryId) => void
}

export function usePaperAutoCancelQuestionEdit(
  args: UsePaperAutoCancelQuestionEditArgs
) {
  const { question, editQuestionFormStates, decUnsavedCount } = args

  function isEditFormStateDirty(
    oldQuestion: Question,
    formState: MergedQuestion
  ): boolean {
    if (!oldQuestion || !formState) return false

    if (
      formState.type !== oldQuestion.type ||
      formState.max_score !== oldQuestion.max_score ||
      formState.correct_answer !== (oldQuestion.correct_answer ?? undefined) ||
      !isEqual(formState.tags, oldQuestion.tags ?? [])
    ) {
      return true
    }

    // Check question data based on type
    if (oldQuestion.type === QuestionType.MCQ) {
      const mcqQuestion = oldQuestion.question
      return (
        formState.question.statement !== mcqQuestion.statement ||
        !isEqual(formState.question.options, mcqQuestion.options)
      )
    }

    const subjectiveQuestion = oldQuestion.question
    return formState.question.statement !== subjectiveQuestion.statement
  }

  watch(question, (_, oldQuestion) => {
    if (!oldQuestion) return

    const formState = editQuestionFormStates[oldQuestion.id]

    // If previous question was in edit mode but not dirty, cancel its edit
    if (formState && !isEditFormStateDirty(oldQuestion, formState)) {
      editQuestionFormStates[oldQuestion.id] = null
      decUnsavedCount(oldQuestion.category_id)
    }
  })
}
