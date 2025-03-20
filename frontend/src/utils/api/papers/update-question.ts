import { QuestionType, type Question } from '~/types'

export interface UpdateMcqQuestionBody
  extends Pick<Question, 'max_score' | 'tags' | 'correct_answer'> {
  type: QuestionType.MCQ
  category_id: number | null
  question: {
    statement: string
    options: string[]
  }
}

export interface UpdateGeneralQuestionBody
  extends Pick<Question, 'max_score' | 'tags' | 'correct_answer'> {
  type: QuestionType.SHORT | QuestionType.LONG
  category_id: number | null
  question: {
    statement: string
  }
}

type UpdateQuestionBody = UpdateMcqQuestionBody | UpdateGeneralQuestionBody

export async function updateQuestion(
  questionId: number,
  paperId: number,
  newData: UpdateQuestionBody
): Promise<void> {
  const { data: question } = useNuxtData<Question>(
    AsyncDataKeys.QUESTION(questionId)
  )

  const previousQuestion: Readonly<Question> = question.value!
  const requestBody: Partial<UpdateQuestionBody> = {}

  // Check each field for changes
  if (newData.type !== previousQuestion.type) {
    requestBody.type = newData.type
    // Must include question data when type changes
    requestBody.question = newData.question
  } else if (
    // Only check question data if type hasn't changed
    newData.type === QuestionType.MCQ &&
    previousQuestion.type === QuestionType.MCQ
  ) {
    const oldQ = previousQuestion.question
    const newQ = newData.question
    if (
      oldQ.statement !== newQ.statement ||
      !arrayEquals(oldQ.options, newQ.options)
    ) {
      requestBody.question = newQ
    }
  } else if (
    previousQuestion.question.statement !== newData.question.statement
  ) {
    requestBody.question = newData.question
  }

  if (newData.max_score !== previousQuestion.max_score) {
    requestBody.max_score = newData.max_score
  }

  if (!arrayEquals(newData.tags, previousQuestion.tags)) {
    requestBody.tags = newData.tags
  }

  // if (newData.category_id !== previousQuestion.category.id) {
  //   requestBody.category_id = newData.category_id
  // }

  if (
    (newData.correct_answer ?? null) !==
    (previousQuestion.correct_answer ?? null)
  ) {
    requestBody.correct_answer = newData.correct_answer
  }

  // If no changes, return early
  if (Object.keys(requestBody).length === 0) return

  question.value = {
    ...question.value!,
    ...newData,
  }

  try {
    await $fetch(`/api/questions/${questionId}`, {
      method: 'PATCH',
      body: requestBody,
      ...getFetchOptions(),
    })

    // Refresh both question and paper since max_score or question_counts might change
    await Promise.all([
      refreshNuxtData(AsyncDataKeys.QUESTION(questionId)),
      refreshNuxtData(AsyncDataKeys.PAPERS_PAPER(paperId)),
    ])
  } catch {
    question.value = previousQuestion
  }
}
