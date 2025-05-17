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
    console.log('getScoreColor: Invalid max score')
    return 'primary' // Prevent division by zero
  }

  const percentage = (scoreObtained / maxScore) * 100

  if (percentage < 40) return 'error'
  else if (percentage < 70) return 'warning'
  else return 'primary'
}
