import { ExamParticipantStatus } from '~/types'

// Middleware to verify if user has participant access to attempt an exam
export default defineNuxtRouteMiddleware(async (to, _from) => {
  try {
    await callOnce(
      async () => {
        const examId = parseInt(to.params.examId as string)

        const res = await checkParticipantAccess(examId)

        if (res.participant_status === ExamParticipantStatus.ENDED) {
          throw res
        }
      },
      { mode: 'navigation' }
    )
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  } catch (err: any) {
    if (err.participant_status === ExamParticipantStatus.ENDED) {
      return abortNavigation({
        statusCode: HttpStatus.FORBIDDEN,
        message: 'You have already attempted this exam',
      })
    }
    if (err.statusCode === HttpStatus.NOT_FOUND) {
      err.message = 'We could not find the exam you are looking for.'
    } else if (err.statusCode === HttpStatus.FORBIDDEN) {
      err.message = 'You are not registered as a participant for this exam.'
    }
    return abortNavigation(err)
  }
})
