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

export const AsyncDataKeys = {
  PAPERS: 'PAPERS',
  PAPERS_PAPER: (paperId: number) => `PAPERS_PAPER_${paperId}`,
  PAPERS_PAPER_QUESTIONS: (paperId: number) =>
    `PAPERS_PAPER_${paperId}_QUESTIONS`,
  PAPERS_PAPER_CATEGORIES: (paperId: number) =>
    `PAPERS_PAPER_${paperId}_CATEGORIES`,
}
