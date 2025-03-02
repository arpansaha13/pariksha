interface LoginBody {
  email: string
  password: string
}

export function login(body: LoginBody) {
  const runtimeConfig = useRuntimeConfig()

  return $fetch(runtimeConfig.public.apiBaseUrl + '/api/auth/login', {
    mode: 'cors',
    method: 'POST',
    body,
  })
}
