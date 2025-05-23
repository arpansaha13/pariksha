interface UpdatePaperTitleBody {
  title?: string
  duration_minutes?: number
}

export async function updatePaper(paperId: number, body: UpdatePaperTitleBody) {
  const { $api } = useNuxtApp()

  const { data: paper } = useNuxtData<Paper>(
    AsyncDataKeys.PAPERS_PAPER(paperId)
  )

  const previousPaper = paper.value!
  paper.value = { ...paper.value!, ...body }

  try {
    await $api<string>(`/api/papers/${paperId}`, {
      method: 'PATCH',
      body,
    })

    await refreshNuxtData(AsyncDataKeys.PAPERS_PAPER(paperId))
  } catch {
    paper.value = previousPaper
  }
}
