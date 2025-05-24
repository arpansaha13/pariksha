export const NUXT_ENV_DEVELOPMENT = 'development'

export const MAX_SCORE_PER_QUESTION = 1000
export const MAX_EXAM_DURATION_MINUTES = 1440 // 24 hours

export const MIN_MCQ_OPTIONS_COUNT = 2
export const MAX_MCQ_OPTIONS_COUNT = 5
export const MAX_CODING_EXAMPLES_COUNT = 4

export const AUTO_SAVE_EXAM_ANSWER_INTERVAL_SECONDS = 15
export const AUTO_SAVE_EVALUATION_INTERVAL_SECONDS = 10

export enum ToastId {
  LOGIN_FAILED = 'login_failed',
  SIGNUP_FAILED = 'signup_failed',
  VERIFY_SIGNUP_FAILED = 'verify_signup_failed',
  FORGOT_PASSWORD_FAILED = 'forgot_password_failed',
  RESET_PASSWORD_FAILED = 'reset_password_failed',
  COPIED_TO_CLIPBOARD = 'copied_to_clipboard',
  INCOMPLETE_EVALUATION = 'incomplete_evaluation',
  DELETE_PAPER_FAILED = 'delete_paper_failed',
  DELETE_EXAM_FAILED = 'delete_exam_failed',
}

export enum HeaderNames {
  XCSRFToken = 'X-CSRFToken',
}

export enum CookieNames {
  CSRF_TOKEN = 'csrftoken',
  TOKEN = 'token',
}

export enum NuxtErrorStatusMessage {
  INCOMPLETE_EVALUATION = 'incomplete_evaluation',
}

export enum HttpStatus {
  OK = 200,
  CREATED = 201,
  NO_CONTENT = 204,
  BAD_REQUEST = 400,
  UNAUTHORIZED = 401,
  FORBIDDEN = 403,
  NOT_FOUND = 404,
  INTERNAL_SERVER_ERROR = 500,
}

export enum UseStateKeys {
  PreviousPath = 'previous-path',
}

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

  EXAM_PARTICIPANT_ANSWERS: (participantId: number) =>
    `PARTICIPANT_${participantId}_ANSWERS`,
} as const

export const InjectionKeys = {
  ConfirmModal: Symbol('ConfirmModal') as InjectionKey<
    ReturnType<ReturnType<typeof useOverlay>['create']>
  >,
}
