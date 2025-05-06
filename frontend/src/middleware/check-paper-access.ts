export default defineNuxtRouteMiddleware(async (to, _from) => {
  if (import.meta.server) {
    const { $api } = useNuxtApp()
    const paperId = parseInt(to.params.paperId as string)

    try {
      await $api(`/api/papers/${paperId}/check`)
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } catch (err: any) {
      if (err.statusCode === HttpStatus.NOT_FOUND) {
        err.message = 'We could not find the paper you are looking for.'
      } else if (err.statusCode === HttpStatus.FORBIDDEN) {
        err.message = 'You do not have access to this paper.'
      }
      return abortNavigation(err)
    }
  }
})
