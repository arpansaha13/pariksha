interface UpdatePaperTitleBody {
  title?: string
  duration_minutes?: number
}

export async function updatePaper(
  paperId: PaperId,
  body: UpdatePaperTitleBody
) {
  const { $api } = useNuxtApp()

  const { data: paper } = useNuxtData<Paper>(UseAsyncDataKeys.paper(paperId))

  const previousPaper = paper.value!
  paper.value = { ...paper.value!, ...body }

  try {
    await $api<string>(`/api/papers/${paperId}`, {
      method: 'PATCH',
      body,
    })

    await refreshNuxtData(UseAsyncDataKeys.paper(paperId))
  } catch {
    paper.value = previousPaper
  }
}
