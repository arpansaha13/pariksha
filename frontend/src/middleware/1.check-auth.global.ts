interface CheckAuthResponse {
  valid: boolean
}

export default defineNuxtRouteMiddleware(async (to, _from) => {
  const isUnprotectedRoute = to.path.startsWith('/auth')
  const { $api } = useNuxtApp()

  if (import.meta.server) {
    const { valid } = await $api<CheckAuthResponse>('/api/check-auth')

    if (valid && isUnprotectedRoute) {
      return navigateTo('/', { replace: true })
    }
    if (!valid && !isUnprotectedRoute) {
      return navigateTo('/auth/login', { replace: true })
    }
  }
})
