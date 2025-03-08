interface CheckAuthResponse {
  valid: boolean
}

export default defineNuxtRouteMiddleware(async (to, _from) => {
  const isUnprotectedRoute = to.path.startsWith('/auth')

  if (import.meta.server) {
    const fetchOptions = getFetchOptions()
    fetchOptions.cache = 'no-cache'

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
