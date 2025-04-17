// Middleware to verify if user has participant access to attempt an exam
export default defineNuxtRouteMiddleware(async (to, _from) => {
  if (import.meta.server) {
    const examId = parseInt(to.params.examId as string)
    const fetchOptions = getFetchOptions()
    fetchOptions.cache = 'no-cache'

    try {
      await $fetch(`/api/exams/${examId}/participants/check`, fetchOptions)
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } catch (err: any) {
      if (err.statusCode === HttpStatus.NOT_FOUND) {
        err.message = 'We could not find the exam you are looking for.'
      } else if (err.statusCode === HttpStatus.FORBIDDEN) {
        err.message = 'You are not registered as a participant for this exam.'
      }
      return abortNavigation(err)
    }
  }
})
