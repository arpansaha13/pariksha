interface CheckAuthResponse {
  valid: boolean
}

export default defineNuxtRouteMiddleware(async (to, _from) => {
  const runtimeConfig = useRuntimeConfig()
  const isUnprotectedRoute = to.path.startsWith('/auth')

  if (import.meta.server) {
    const fetchOptions: Parameters<typeof $fetch>[1] = {
      cache: 'no-cache',
      baseURL: runtimeConfig.apiBaseUrl,
    }

    const csrftoken = useCookie('csrftoken')
    const token = useCookie('token') // Server-side JS can read HttpOnly cookies

    if (csrftoken.value) {
      fetchOptions.headers = {
        'X-CSRFToken': csrftoken.value,

        // `credentials: include` won't work in server-side
        // because it is the browser that appends the cookies to request headers
        'Cookie': `token=${token.value}`,
      }
    }

    const { valid } = await $fetch<CheckAuthResponse>(
      '/api/check-auth',
      fetchOptions
    )

    if (valid && isUnprotectedRoute) {
      return navigateTo('/')
    }
    if (!valid && !isUnprotectedRoute) {
      return navigateTo('/auth/login')
    }
  }
})
