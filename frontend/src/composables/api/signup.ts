interface SignUpBody {
  email: string
  password: string
}

export function signUp(body: SignUpBody) {
  const runtimeConfig = useRuntimeConfig()

  return $fetch(runtimeConfig.public.apiBaseUrl + '/api/auth/signup', {
    mode: 'cors',
    method: 'POST',
    body,
  })
}
