interface LoginBody {
  email: string
  password: string
}

export function login(body: LoginBody) {
  return $fetch('/api/auth/login', {
    method: 'POST',
    body,
  })
}
