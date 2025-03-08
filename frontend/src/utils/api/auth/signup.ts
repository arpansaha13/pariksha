interface SignUpBody {
  email: string
  password: string
}

export function signUp(body: SignUpBody) {
  return $fetch('/api/auth/signup', {
    method: 'POST',
    body,
  })
}
