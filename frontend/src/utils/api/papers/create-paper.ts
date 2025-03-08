enum PaperOwnership {
  OWNER = 'OWNER',
  SHARED = 'SHARED',
}

interface PaperQuestionCounts {
  mcq: number
  short: number
  long: number
}

interface CreatePaperResponse {
  id: number
  title: number
  maxScore: number
  questionCounts: PaperQuestionCounts
  paperOwnership: {
    id: number
    path: string
    type: PaperOwnership
  }
}

export async function createPaper() {
  const csrftoken = useCookie(CookieNames.CSRF_TOKEN)

  const res = await $fetch<string>('/api/papers', {
    method: 'POST',
    credentials: 'include',
    headers: {
      [HeaderNames.XCSRFToken]: csrftoken.value!,
    },
  })

  return JSON.parse(res) as CreatePaperResponse
}
