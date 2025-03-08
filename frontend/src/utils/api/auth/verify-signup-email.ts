interface VerifySignUpEmailBody {
  email: string
  otp: string
}

export function verifySignUpEmail(body: VerifySignUpEmailBody) {
  return $fetch('/api/auth/verification/signup', {
    method: 'POST',
    body,
  })
}
