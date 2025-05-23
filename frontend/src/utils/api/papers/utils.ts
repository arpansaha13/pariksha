export type MergedQuestionOmit = 'id' | 'paper_id' | 'order'

export interface MergedQuestion
  extends Omit<BaseQuestion, MergedQuestionOmit | 'category_id'> {
  type: QuestionType
  question: QuestionMcqContent &
    QuestionSubjectiveContent &
    QuestionCodingContent
}

export const extractQuestionContent = (
  mergedQuestion: MergedQuestion
):
  | QuestionMcqContent
  | QuestionSubjectiveContent
  | QuestionCodingContent
  | null => {
  if (mergedQuestion.type === QuestionType.MCQ) {
    return extractMcqQuestionContent(mergedQuestion)
  }
  if (mergedQuestion.type === QuestionType.SUBJECTIVE) {
    return extractSubjectiveQuestionContent(mergedQuestion)
  }
  if (mergedQuestion.type === QuestionType.CODING) {
    return extractCodingQuestionContent(mergedQuestion)
  }

  logWarning(`Invalid mergedQuestion type: ${mergedQuestion.type}`)
  return null
}

const extractMcqQuestionContent = (
  mergedQuestion: MergedQuestion
): QuestionMcqContent => {
  return {
    statement: mergedQuestion.question.statement,
    options: mergedQuestion.question.options,
  }
}

const extractSubjectiveQuestionContent = (
  mergedQuestion: MergedQuestion
): QuestionSubjectiveContent => {
  return {
    statement: mergedQuestion.question.statement,
  }
}

const extractCodingQuestionContent = (
  mergedQuestion: MergedQuestion
): QuestionCodingContent => {
  const examples = mergedQuestion.question.examples?.filter(example => {
    // Skip if all fields are empty
    if (!example.input && !example.output && !example.explanation) {
      return false
    }

    // Validate input/output pair
    if (
      (example.input && !example.output) ||
      (!example.input && example.output)
    ) {
      logWarning('Example must have both input and output or neither')
      return false
    }

    return true
  })

  return {
    title: mergedQuestion.question.title,
    statement: mergedQuestion.question.statement,
    examples: examples?.length ? examples : undefined,
  }
}
