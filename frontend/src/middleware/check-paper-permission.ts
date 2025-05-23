export default defineNuxtRouteMiddleware(async (to, _from) => {
  try {
    await callOnce(
      async () => {
        const paperId = to.params.paperId as PaperId
        const { data, error, status } = await usePaperPermission(paperId)
        if (status.value === 'error') {
          throw error.value
        }
        if (!data.value!.can_read) {
          throw createError({ statusCode: HttpStatus.FORBIDDEN })
        }
      },
      { mode: 'navigation' }
    )
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  } catch (err: any) {
    if (err.statusCode === HttpStatus.NOT_FOUND) {
      err.message = 'We could not find the paper you are looking for.'
    } else if (err.statusCode === HttpStatus.FORBIDDEN) {
      err.message = 'You do not have access to this paper.'
    }
    return abortNavigation(err)
  }
})
