import { ExamPermission } from '~/types/exam'

export default defineNuxtRouteMiddleware(async (to, _from) => {
  try {
    await callOnce(
      async () => {
        const examId = parseInt(to.params.examId as string)

        const { data } = await useExamCheckAccess(examId)

        if (data.value!.access_type === ExamPermission.PARTICIPANT) {
          setPageLayout('blank')
        }
      },
      { mode: 'navigation' }
    )
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  } catch (err: any) {
    if (err.statusCode === HttpStatus.NOT_FOUND) {
      err.message = 'We could not find the exam you are looking for.'
    } else if (err.statusCode === HttpStatus.FORBIDDEN) {
      err.message = 'You do not have access to this exam.'
    }
    return abortNavigation(err)
  }
})
