interface ForgotPasswordBody {
  email: string
}

export function forgotPassword(body: ForgotPasswordBody) {
  return $fetch('/api/auth/forgot-password', {
    method: 'POST',
    body,
  })
}
