export const AsyncDataKeys = {
  PAPERS: 'PAPERS',
  PAPERS_PAPER: (paperId: PaperId) => `PAPERS_PAPER_${paperId}`,
  PAPERS_PAPER_QUESTIONS: (paperId: PaperId) =>
    `PAPERS_PAPER_${paperId}_QUESTIONS`,
  PAPERS_PAPER_CATEGORIES: (paperId: PaperId) =>
    `PAPERS_PAPER_${paperId}_CATEGORIES`,
  PAPER_PERMISSION: (paperId: PaperId) => `PAPER_${paperId}_PERMISSION`,

  QUESTION: (questionId: number | null) =>
    questionId ? `QUESTION_${questionId}` : 'QUESTION',

  EXAMS: 'EXAMS',
  EXAM: (examId: ExamId) => `EXAM_${examId}`,
  EXAM_PERMISSION: (examId: ExamId) => `EXAM_${examId}_PERMISSION`,
  EXAM_PARTICIPANT: (examId: ExamId) => `EXAM_${examId}_PARTICIPANT`,
  EXAM_PARTICIPANTS: (examId: ExamId) => `EXAM_${examId}_PARTICIPANTS`,
  EXAM_QUESTIONS: (examId: ExamId) => `EXAM_${examId}_QUESTIONS`,
  EXAM_CATEGORIES: (examId: ExamId) => `EXAM_${examId}_CATEGORIES`,
  EXAM_RESULTS: (examId: ExamId) => `EXAM_${examId}_RESULTS`,

  EXAM_QUESTION: (questionId: number | null) =>
    questionId ? `EXAM_QUESTION_${questionId}` : 'EXAM_QUESTION',

  EXAM_PARTICIPANT_BY_ID: (participantId: ExamParticipantId) =>
    `PARTICIPANT_${participantId}`,

  EVALUATION_ANSWER: (participantId: number, questionId: number | null) =>
    participantId && questionId
      ? `PARTICIPANT_${participantId}_QUESTION_${questionId}_EVALUATION_ANSWER`
      : 'EVALUATION_ANSWER',
  ANSWER_EVALUATION_DATA: (participantId: number, questionId: number | null) =>
    participantId && questionId
      ? `PARTICIPANT_${participantId}_QUESTION_${questionId}_ANSWER_EVALUATION_DATA`
      : 'ANSWER_EVALUATION_DATA',

  AUTH_USER: 'AUTH_USER',

  BOILERPLATE: (questionId: QuestionId | null, languageId: LanguageId | null) =>
    `BOILERPLATE_${questionId}_${languageId}`,

  EXAM_PARTICIPANT_ANSWERS: (participantId: number) =>
    `PARTICIPANT_${participantId}_ANSWERS`,
} as const
