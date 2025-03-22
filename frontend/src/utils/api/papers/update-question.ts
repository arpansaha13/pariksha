import { isNullOrUndefined } from '@arpansaha13/utils'
import { QuestionType, type Question, type QuestionMinimal } from '~/types'

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
  const { data: groupedQuestions } = useNuxtData<
    Record<number, QuestionMinimal[]>
  >(AsyncDataKeys.PAPERS_PAPER_QUESTIONS(paperId))

  const previousQuestion = question.value!
  const requestBody = getRequestBody(previousQuestion, newData)

  // If no changes, return early
  if (Object.keys(requestBody).length === 0) return

  // Store minimal data for rollback
  const rollbackData = {
    id: questionId,
    categoryId: previousQuestion.category.id,
    question: previousQuestion.question,
  }

  // Optimistically update the full question
  question.value = {
    ...question.value!,
    ...newData,
  }

  // Update the particular question in paperQuestions
  const categoryQuestions = groupedQuestions.value![rollbackData.categoryId]
  const minimalQuestion = categoryQuestions?.find(q => q.id === questionId)
  if (minimalQuestion && requestBody.question) {
    minimalQuestion.question = requestBody.question
  }

  try {
    await $fetch(`/api/questions/${questionId}`, {
      method: 'PATCH',
      body: requestBody,
      ...getFetchOptions(),
    })

    const refreshPromises = [
      refreshNuxtData(AsyncDataKeys.QUESTION(questionId)),
    ]

    const isMaxScoreUpdated =
      !isNullOrUndefined(requestBody.max_score) &&
      requestBody.max_score !== previousQuestion.max_score
    const isTypeUpdated =
      !isNullOrUndefined(requestBody.type) &&
      requestBody.type !== previousQuestion.type

    // Refresh paper data if max_score or type changed
    if (isMaxScoreUpdated || isTypeUpdated) {
      refreshPromises.push(refreshNuxtData(AsyncDataKeys.PAPERS_PAPER(paperId)))
    }

    await Promise.all(refreshPromises)
  } catch {
    // The active question may change during the fetch
    // Rollback question only if IDs match
    if (question.value?.id === previousQuestion.id) {
      question.value = previousQuestion
    }

    // Rollback the question in paperQuestions
    const minimalQuestion = groupedQuestions.value![
      rollbackData.categoryId
    ]?.find(q => q.id === rollbackData.id)
    if (minimalQuestion) {
      minimalQuestion.question = rollbackData.question
    }
  }
}

/** Only include fields that are updated */

function getRequestBody(
  previousQuestion: Readonly<Question>,
  newData: UpdateQuestionBody
) {
  const requestBody: Partial<UpdateQuestionBody> = {}

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

  return requestBody
}
