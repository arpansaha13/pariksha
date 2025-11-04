import { isNullOrUndefined } from '@arpansaha13/utils'
import { isString as _isString } from 'lodash-es'

/**
 * Checks if a string contains only alphabetic characters (A-Z, a-z)
 * @param str - The string to check
 * @returns boolean - True if string contains only alphabets, false otherwise
 */
export const isAlpha = (str: string): boolean => {
  return _isString(str) && /^[A-Za-z]+$/.test(str)
}

type ScoreColor = 'error' | 'warning' | 'primary'

/**
 * Determines the color based on the score percentage.
 *
 * @param scoreObtained - The score obtained by the student.
 * @param maxScore - The maximum possible score.
 * @returns The color representing the score range.
 */
export function getScoreColor(
  scoreObtained: number,
  maxScore: number
): ScoreColor {
  if (maxScore <= 0) {
    logWarning('Invalid max score')
    return 'primary' // Prevent division by zero
  }

  const percentage = (scoreObtained / maxScore) * 100

  if (percentage < 40) return 'error'
  else if (percentage < 70) return 'warning'
  else return 'primary'
}

export function formatDurationMinutes(durationMinutes: number) {
  if (isNullOrUndefined(durationMinutes)) return '0 minutes'

  const hours = calcHours(durationMinutes)
  const remainingMinutes = calcRemainderMinutes(durationMinutes)

  if (hours === 0) return `${remainingMinutes} minutes`
  if (remainingMinutes === 0)
    return `${hours} ${hours === 1 ? 'hour' : 'hours'}`
  if (hours === 1)
    return `${hours} hour ${remainingMinutes} ${remainingMinutes === 1 ? 'minute' : 'minutes'}`
  return `${hours} hours  ${remainingMinutes} minutes`
}

/**
 * Returns the label for a given input definition for a coding question.
 * If the value is `QuestionCodingContentInputTypes.ARRAY`, it checks the first item's type in `items`
 * and returns `"Array of [type]"`, where `[type]` is derived from the enum mapping.
 *
 * @param {QuestionCodingContentInputTypes} value - The enum value to convert to a string.
 * @param {Array<{ type: QuestionCodingContentInputTypes }>} [items] - Optional array of objects containing type information.
 * @returns {string} The formatted string representation of the enum value.
 */
export function getCodingQuestionParameterLabel(
  value: QuestionCodingContentInputTypes,
  items?: { type: QuestionCodingContentPrimitiveInputTypes }[]
): string {
  const enumMappings: Record<QuestionCodingContentInputTypes, string> = {
    [QuestionCodingContentPrimitiveInputTypes.NUMBER]: 'Number',
    [QuestionCodingContentPrimitiveInputTypes.STRING]: 'String',
    [QuestionCodingContentPrimitiveInputTypes.BOOLEAN]: 'Boolean',
    [QuestionCodingContentCompositeInputTypes.ARRAY]: 'Array',
    // [QuestionCodingContentCompositeInputTypes.OBJECT]: 'Object',
    // [QuestionCodingContentCompositeInputTypes.TUPLE]: 'Tuple',
  }

  const pluralize = (word: string) => `${word}s` as const

  // Special handling for ARRAY type
  if (
    value === QuestionCodingContentCompositeInputTypes.ARRAY &&
    items?.[0].type
  ) {
    const firstItemType = items[0].type
    return `Array of ${pluralize(enumMappings[firstItemType].toLowerCase() || 'unknown')}`
  }

  return enumMappings[value] || 'Unknown'
}

export function getDefaultCodingQuestionInputVariableName(serial: number) {
  return `arg${serial}` as const
}
