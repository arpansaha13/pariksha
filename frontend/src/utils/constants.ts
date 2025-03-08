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
