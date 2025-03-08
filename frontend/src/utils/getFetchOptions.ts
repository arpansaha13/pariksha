/**
 * Adds the csrftoken to request headers.
 * If `import.meta.server = true`, then it will also append session-cookie to headers.
 */

export function getFetchOptions() {
  const csrftoken = useCookie(CookieNames.CSRF_TOKEN)
  const fetchOptions: Parameters<typeof $fetch>[1] = {}

  if (csrftoken.value) {
    fetchOptions.headers = {
      [HeaderNames.XCSRFToken]: csrftoken.value,
    }
  }

  if (import.meta.client) {
    fetchOptions.credentials = 'include'
  } else {
    const runtimeConfig = useRuntimeConfig()
    fetchOptions.baseURL = runtimeConfig.apiBaseUrl

    const token = useCookie(CookieNames.TOKEN) // Server-side JS can read HttpOnly cookies
    if (token.value) {
      // https://nuxt.com/docs/api/utils/dollarfetch#passing-headers-and-cookies
      fetchOptions.headers = {
        ...fetchOptions.headers,
        Cookie: `token=${token.value}`,
      }
    }
  }

  return fetchOptions
}
