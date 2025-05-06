export default defineNuxtPlugin(nuxtApp => {
  const csrftoken = useCookie(CookieNames.CSRF_TOKEN)

  const api = $fetch.create({
    onRequest({ options }) {
      if (csrftoken.value) {
        options.headers.set(HeaderNames.XCSRFToken, csrftoken.value)
      }

      if (import.meta.client) {
        options.credentials = 'include'
      } else {
        const runtimeConfig = useRuntimeConfig()
        options.baseURL = runtimeConfig.apiBaseUrl

        const token = useCookie(CookieNames.TOKEN) // Server-side JS can read HttpOnly cookies
        if (token.value) {
          // https://nuxt.com/docs/api/utils/dollarfetch#passing-headers-and-cookies
          options.headers.set('Cookie', `token=${token.value}`)
        }
      }
    },
    async onResponseError({ response }) {
      const route = useRoute()
      const isProtectedRoute = !route.path.startsWith('/auth')

      if (response.status === 401 && isProtectedRoute) {
        await nuxtApp.runWithContext(() => navigateTo('/auth/login'))
      }
    },
  })

  // Expose to useNuxtApp().$api
  return {
    provide: {
      api,
    },
  }
})
