import type { Question, QuestionType } from '~/types'

export interface CreateMcqQuestionBody {
  type: QuestionType.MCQ
  category_id: number | null
  question: {
    statement: string
    options: string[]
  }
  max_score: number
  tags: string[]
  correct_answer: string | null | undefined
}

export interface CreateGeneralQuestionBody {
  type: QuestionType.SHORT | QuestionType.LONG
  category_id: number | null
  question: {
    statement: string
  }
  max_score: number
  tags: string[]
  correct_answer: string | null | undefined
}

type CreateQuestionBody = CreateMcqQuestionBody | CreateGeneralQuestionBody

export async function createQuestion(
  paperId: number,
  body: CreateQuestionBody
): Promise<void> {
  await $fetch<Question>(`/api/papers/${paperId}/questions`, {
    method: 'POST',
    body,
    ...getFetchOptions(),
  })

  // Refresh both questions and paper data since max_score and question_counts change
  await Promise.all([
    refreshNuxtData(AsyncDataKeys.PAPERS_PAPER_QUESTIONS(paperId)),
    refreshNuxtData(AsyncDataKeys.PAPERS_PAPER(paperId)),
  ])
}
