import { z } from 'zod'
import {
  QuestionType,
  QuestionCodingContentPrimitiveInputTypes,
  QuestionCodingContentCompositeInputTypes,
} from '../types/models'

// Common question fields
const baseQuestionSchema = z.object({
  id: z.number().int(),
  order: z.number().int(),
  category_id: z.number().int(),
  tags: z.array(z.string()),
  paper_id: z.number().int(),
  max_score: z.number().int().min(0).max(1000),
  correct_answer: z.string().optional(),
})

// MCQ Question
const mcqQuestionContentSchema = z.object({
  statement: z.string().min(1),
  options: z.array(z.string()).min(2).max(5),
})

const mcqQuestionSchema = baseQuestionSchema.extend({
  type: z.literal(QuestionType.MCQ),
  question: mcqQuestionContentSchema,
})

// Subjective Question
const subjectiveQuestionContentSchema = z.object({
  statement: z.string().min(1),
})

const subjectiveQuestionSchema = baseQuestionSchema.extend({
  type: z.literal(QuestionType.SUBJECTIVE),
  question: subjectiveQuestionContentSchema,
})

// Coding Question
const parameterItemSchema = z.object({
  property_name: z.string().optional(),
  type: z.nativeEnum(QuestionCodingContentPrimitiveInputTypes),
})

const parameterSchema = z.object({
  type: z.union([
    z.nativeEnum(QuestionCodingContentPrimitiveInputTypes),
    z.nativeEnum(QuestionCodingContentCompositeInputTypes),
  ]),
  items: z.array(parameterItemSchema).optional(),
})

const inputDefinitionSchema = parameterSchema.extend({
  variable_name: z.string().min(1),
})

const testCaseSchema = z.object({
  inputs: z.array(z.string()),
  output: z.string(),
  explanation: z.string().optional(),
})

const codingQuestionContentSchema = z.object({
  title: z.string().min(1),
  statement: z.string().min(1),
  input_definitions: z.array(inputDefinitionSchema).min(1),
  output_definition: parameterSchema,
  test_cases: z.array(testCaseSchema).optional(),
})

const codingQuestionSchema = baseQuestionSchema.extend({
  type: z.literal(QuestionType.CODING),
  question: codingQuestionContentSchema,
})

// Combined Question schema
export const questionSchema = z.discriminatedUnion(
  'type',
  [mcqQuestionSchema, subjectiveQuestionSchema, codingQuestionSchema],
  {
    invalid_type_error: 'Please select the type of question.',
    required_error: 'Please select the type of question.',
  }
)

export type QuestionSchema = z.infer<typeof questionSchema>

// Helper schemas
export const mcqAnswerSchema = z.object({
  optionIndex: z.number().int().min(0).optional(),
})

export const subjectiveAnswerSchema = z.object({
  text: z.string(),
})

export const answerSchema = z.object({
  id: z.number().int(),
  exam_participant_id: z.number().int(),
  question_id: z.number().int(),
  answer: z.union([mcqAnswerSchema, subjectiveAnswerSchema, z.null()]),
  score_awarded: z.number(),
  comments: z.string(),
})
