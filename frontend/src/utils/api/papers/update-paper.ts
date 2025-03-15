import type { Paper } from '~/types'

interface UpdatePaperTitleBody {
  title: string
}

export async function updatePaper(paperId: number, body: UpdatePaperTitleBody) {
  const { data: paper } = useNuxtData<Paper>(
    AsyncDataKeys.PAPERS_PAPER(paperId)
  )

  const previousPaper = paper.value!
  paper.value = { ...paper.value!, ...body }

  try {
    await $fetch<string>(`/api/papers/${paperId}`, {
      method: 'PATCH',
      body,
      ...getFetchOptions(),
    })

    await refreshNuxtData(AsyncDataKeys.PAPERS_PAPER(paperId))
  } catch {
    paper.value = previousPaper
  }
}
