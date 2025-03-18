import type { Question, QuestionType } from '~/types'

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
  body: UpdateQuestionBody
): Promise<void> {
  const { data: question } = useNuxtData<Question>(
    AsyncDataKeys.QUESTION(questionId)
  )

  const previousQuestion = question.value

  question.value = {
    ...question.value!,
    ...body,
  }

  try {
    await $fetch(`/api/questions/${questionId}`, {
      method: 'PATCH',
      body,
      ...getFetchOptions(),
    })

    // Refresh both questions and paper data since max_score or question_counts might change
    await Promise.all([
      refreshNuxtData(AsyncDataKeys.QUESTION(questionId)),
      refreshNuxtData(AsyncDataKeys.PAPERS_PAPER(paperId)),
    ])
  } catch {
    question.value = previousQuestion
  }
}
