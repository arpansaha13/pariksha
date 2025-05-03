import { isNullOrUndefined } from '@arpansaha13/utils'
import type { Question, QuestionMcq, QuestionMinimal } from '~/types'

type UpdateQuestionBody = Partial<Question>

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
    categoryId: previousQuestion.category_id,
    question: previousQuestion.question,
  }

  // Optimistically update the full question
  question.value = {
    ...question.value!,
    ...(newData as (typeof question)['value']),
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

    // requestBody will only have the updated fields
    const isMaxScoreUpdated = !isNullOrUndefined(requestBody.max_score)
    const isTypeUpdated = !isNullOrUndefined(requestBody.type)

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
  const requestBody: UpdateQuestionBody = {}

  /** Compares options assuming MCQ question */
  const areOpionsEqual = (a: Question['question'], b: Question['question']) => {
    return arrayEquals(
      (a as QuestionMcq['question']).options ?? [],
      (b as QuestionMcq['question']).options ?? []
    )
  }

  if (
    !isNullOrUndefined(newData.type) &&
    newData.type !== previousQuestion.type
  ) {
    requestBody.type = newData.type
    if (isNullOrUndefined(newData.question)) {
      throw new Error('Must include question data when type changes')
    }
  }

  if (!isNullOrUndefined(newData.question)) {
    const oldQ = previousQuestion.question
    const newQ = newData.question
    console.log(oldQ)
    console.log(newQ)
    if (oldQ.statement !== newQ.statement || !areOpionsEqual(newQ, oldQ)) {
      requestBody.question = newQ
    }
  }

  if (
    !isNullOrUndefined(newData.max_score) &&
    newData.max_score !== previousQuestion.max_score
  ) {
    requestBody.max_score = newData.max_score
  }

  if (
    !isNullOrUndefined(newData.tags) &&
    !arrayEquals(newData.tags, previousQuestion.tags)
  ) {
    requestBody.tags = newData.tags
  }

  // if (newData.category_id !== previousQuestion.category.id) {
  //   requestBody.category_id = newData.category_id
  // }

  if (
    !isNullOrUndefined(newData.correct_answer) &&
    newData.correct_answer !== previousQuestion.correct_answer
  ) {
    requestBody.correct_answer = newData.correct_answer
  }

  return requestBody
}
