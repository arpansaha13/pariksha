export type MergedQuestionOmit = 'id' | 'paper_id' | 'order' | 'category_id'

export interface MergedQuestion extends Omit<BaseQuestion, MergedQuestionOmit> {
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
  const testCases = mergedQuestion.question.test_cases?.filter(testCase => {
    // Skip if all fields are empty
    if (!testCase.inputs && !testCase.output && !testCase.explanation) {
      return false
    }

    // Validate input/output pair
    if (
      (testCase.inputs && !testCase.output) ||
      (!testCase.inputs && testCase.output)
    ) {
      logWarning('Example must have both input and output or neither')
      return false
    }

    return true
  })

  return {
    title: mergedQuestion.question.title,
    statement: mergedQuestion.question.statement,
    test_cases: testCases?.length ? testCases : undefined,
    input_definitions: mergedQuestion.question.input_definitions,
    output_definition: mergedQuestion.question.output_definition,
  }
}
