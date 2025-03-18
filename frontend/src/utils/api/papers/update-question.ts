import type { Question, QuestionType } from '~/types'

export interface UpdateMcqQuestionBody {
  type?: QuestionType.MCQ
  category_id?: number | null
  question?: {
    statement: string
    options: string[]
  }
  max_score?: number
  tags?: string[]
  correct_answer?: string
}

export interface UpdateGeneralQuestionBody {
  type?: QuestionType.SHORT | QuestionType.LONG
  category_id?: number | null
  question?: {
    statement: string
  }
  max_score?: number
  tags?: string[]
  correct_answer?: string
}

type UpdateQuestionBody = UpdateMcqQuestionBody | UpdateGeneralQuestionBody

export async function updateQuestion(
  questionId: number,
  paperId: number,
  body: UpdateQuestionBody
): Promise<void> {
  const { data: questions } = useNuxtData<Record<number, Question[]>>(
    AsyncDataKeys.PAPERS_PAPER_QUESTIONS(paperId)
  )

  const previousQuestions = JSON.parse(JSON.stringify(questions.value))

  for (const categoryQuestions of Object.values(questions.value!)) {
    const questionToUpdate = categoryQuestions.find(q => q.id === questionId)
    if (questionToUpdate) {
      Object.assign(questionToUpdate, {
        ...questionToUpdate,
        ...body,
        question: body.question
          ? { ...questionToUpdate.question, ...body.question }
          : questionToUpdate.question,
      })
      break
    }
  }

  try {
    await $fetch(`/api/questions/${questionId}`, {
      method: 'PATCH',
      body,
      ...getFetchOptions(),
    })

    // Refresh both questions and paper data since max_score or question_counts might change
    await Promise.all([
      refreshNuxtData(AsyncDataKeys.PAPERS_PAPER_QUESTIONS(paperId)),
      refreshNuxtData(AsyncDataKeys.PAPERS_PAPER(paperId)),
    ])
  } catch {
    questions.value = previousQuestions
  }
}
