import type { Question, QuestionType } from '~/types'

export interface CreateMcqQuestionBody {
  type: QuestionType.MCQ
  category_id: number
  question: {
    statement: string
    options: string[]
  }
  max_score: number
  tags: string[]
  correct_answer: string | null | undefined
}

export interface CreateSubjectiveQuestionBody {
  type: QuestionType.SUBJECTIVE
  category_id: number
  question: {
    statement: string
  }
  max_score: number
  tags: string[]
  correct_answer: string | null | undefined
}

type CreateQuestionBody = CreateMcqQuestionBody | CreateSubjectiveQuestionBody

export async function createQuestion(
  paperId: number,
  body: CreateQuestionBody
): Promise<void> {
  const { $api } = useNuxtApp()

  await $api<Question>(`/api/papers/${paperId}/questions`, {
    method: 'POST',
    body,
  })

  // Refresh both questions and paper data since max_score and question_counts change
  await Promise.all([
    refreshNuxtData(AsyncDataKeys.PAPERS_PAPER_QUESTIONS(paperId)),
    refreshNuxtData(AsyncDataKeys.PAPERS_PAPER(paperId)),
  ])
}
