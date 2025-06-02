import {
  extractQuestionContent,
  type MergedQuestion,
  type MergedQuestionOmit,
} from './utils'

type CreateQuestionOmit = Exclude<MergedQuestionOmit, 'category_id'>

type CreateQuestionBody =
  | Omit<QuestionMcq, CreateQuestionOmit>
  | Omit<QuestionSubjective, CreateQuestionOmit>
  | Omit<QuestionCoding, CreateQuestionOmit>

type CreateQuestionReturn = Pick<Question, 'id'>

export async function createQuestion(
  paperId: PaperId,
  categoryId: CategoryId,
  mergedQuestion: MergedQuestion
): Promise<number | null> {
  const { $api } = useNuxtApp()

  const body: CreateQuestionBody = {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    type: mergedQuestion.type as any,
    max_score: mergedQuestion.max_score,
    tags: mergedQuestion.tags,
    // correct_answer: mergedQuestion.correct_answer,
    category_id: categoryId,
    question: extractQuestionContent(mergedQuestion)!,
  }

  const res = await $api<string | CreateQuestionReturn>(
    `/api/papers/${paperId}/questions`,
    {
      method: 'POST',
      body,
    }
  )

  const parsedRes = typeof res === 'string' ? JSON.parse(res) : res

  // Refresh both questions and paper data since max_score and question_counts change
  await Promise.all([
    refreshNuxtData(UseAsyncDataKeys.paper_questions(paperId)),
    refreshNuxtData(UseAsyncDataKeys.paper(paperId)),
  ])

  return parsedRes.id
}
