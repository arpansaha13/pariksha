import {
  extractQuestionContent,
  type MergedQuestion,
  type MergedQuestionOmit,
} from './utils'

type CreateQuestionBody =
  | Omit<QuestionMcq, MergedQuestionOmit>
  | Omit<QuestionSubjective, MergedQuestionOmit>
  | Omit<QuestionCoding, MergedQuestionOmit>

type CreateQuestionReturn = Pick<Question, 'id'>

export async function createQuestion(
  paperId: number,
  categoryId: number,
  mergedQuestion: MergedQuestion
): Promise<number | null> {
  const { $api } = useNuxtApp()

  const body = {
    type: mergedQuestion.type,
    max_score: mergedQuestion.max_score,
    tags: mergedQuestion.tags,
    correct_answer: mergedQuestion.correct_answer,
    category_id: categoryId,
    question: extractQuestionContent(mergedQuestion)!,
  } as CreateQuestionBody

  const res = await $api<CreateQuestionReturn>(
    `/api/papers/${paperId}/questions`,
    {
      method: 'POST',
      body,
    }
  )

  // Refresh both questions and paper data since max_score and question_counts change
  await Promise.all([
    refreshNuxtData(AsyncDataKeys.PAPERS_PAPER_QUESTIONS(paperId)),
    refreshNuxtData(AsyncDataKeys.PAPERS_PAPER(paperId)),
  ])

  return res.id
}
