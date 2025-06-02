export const UseAsyncDataKeys = {
  auth_user: 'auth_user',

  papers: 'papers',
  paper: (paperId: PaperId) => `paper_${paperId}`,
  paper_questions: (paperId: PaperId) => `paper_${paperId}_questions`,
  paper_categories: (paperId: PaperId) => `paper_${paperId}_categories`,
  paper_permission: (paperId: PaperId) => `paper_${paperId}_permission`,

  exams: 'exams',
  exam: (examId: ExamId) => `exam_${examId}`,
  exam_permission: (examId: ExamId) => `exam_${examId}_permission`,
  exam_participant: (examId: ExamId) => `exam_${examId}_participant`,
  exam_participants: (examId: ExamId) => `exam_${examId}_participants`,
  exam_questions: (examId: ExamId) => `exam_${examId}_questions`,
  exam_categories: (examId: ExamId) => `exam_${examId}_categories`,
  exam_results: (examId: ExamId) => `exam_${examId}_results`,

  paper_question: (questionId: QuestionId | null) =>
    `paper_question_${questionId}`,
  exam_question: (questionId: QuestionId | null) =>
    `exam_question_${questionId}`,

  participant_by_id: (participantId: ExamParticipantId) =>
    `participant_${participantId}`,

  evaluation_answer: (participantId: number, questionId: number | null) =>
    `participant_${participantId}_question_${questionId}_evaluation_answer`,
  answer_evaluation_data: (participantId: number, questionId: number | null) =>
    `participant_${participantId}_question_${questionId}_answer_evaluation_data`,

  boilerplate: (questionId: QuestionId | null, languageId: LanguageId | null) =>
    `boilerplate_${questionId}_${languageId}`,

  exam_participant_answers: (participantId: number) =>
    `participant_${participantId}_answers`,
}
