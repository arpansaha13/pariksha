interface CheckAuthResponse {
  valid: boolean
}

export default defineNuxtRouteMiddleware(async (to, _from) => {
  const runtimeConfig = useRuntimeConfig()
  const isUnprotectedRoute = to.path.startsWith('/auth')

  if (import.meta.server) {
    const { valid } = await $fetch<CheckAuthResponse>(
      runtimeConfig.apiBaseUrl + '/api/check-auth',
      {
        mode: 'cors',
        credentials: 'include',
      }
    )

    if (valid && isUnprotectedRoute) {
      return navigateTo('/')
    }
    if (!valid && !isUnprotectedRoute) {
      return navigateTo('/auth/login')
    }
  }
})
