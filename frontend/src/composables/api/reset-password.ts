interface ResetPasswordBody {
  email: string
  new_password: string
  otp: string
}

export function resetPassword(body: ResetPasswordBody) {
  return $fetch('/api/auth/reset-password', {
    method: 'POST',
    body,
  })
}
