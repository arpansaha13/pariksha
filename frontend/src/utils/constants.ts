export enum ToastId {
  LOGIN_FAILED = 'login_failed',
  SIGNUP_FAILED = 'signup_failed',
  VERIFY_SIGNUP_FAILED = 'verify_signup_failed',
  FORGOT_PASSWORD_FAILED = 'forgot_password_failed',
  RESET_PASSWORD_FAILED = 'reset_password_failed',
}

export enum HeaderNames {
  XCSRFToken = 'X-CSRFToken',
}

export enum CookieNames {
  CSRF_TOKEN = 'csrftoken',
  TOKEN = 'token',
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

export const AsyncDataKeys = {
  PAPERS: 'PAPERS',
  PAPERS_PAPER: (paperId: number) => `PAPERS_PAPER_${paperId}`,
  PAPERS_PAPER_QUESTIONS: (paperId: number) =>
    `PAPERS_PAPER_${paperId}_QUESTIONS`,
  PAPERS_PAPER_CATEGORIES: (paperId: number) =>
    `PAPERS_PAPER_${paperId}_CATEGORIES`,

  QUESTION: (questionId?: number | null) =>
    questionId ? `QUESTION_${questionId}` : 'QUESTION',

  EXAMS: 'EXAMS',
  EXAM: (examId: number) => `EXAM_${examId}`,
  EXAM_PARTICIPANT: (examId: number) => `EXAM_${examId}_PARTICIPANT`,
  EXAM_QUESTIONS: (examId: number) => `EXAM_${examId}_QUESTIONS`,
  EXAM_CATEGORIES: (examId: number) => `EXAM_${examId}_CATEGORIES`,

  EXAM_QUESTION: (questionId?: number | null) =>
    questionId ? `EXAM_QUESTION_${questionId}` : 'EXAM_QUESTION',
}

export const InjectionKeys = {
  ConfirmModal: Symbol('ConfirmModal') as InjectionKey<
    ReturnType<ReturnType<typeof useOverlay>['create']>
  >,
}
