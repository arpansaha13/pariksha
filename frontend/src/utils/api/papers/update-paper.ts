interface UpdatePaperTitleBody {
  title: string
}

export async function updatePaper(paperId: number, body: UpdatePaperTitleBody) {
  await $fetch(`/api/papers/${paperId}`, {
    method: 'PATCH',
    body,
    ...getFetchOptions(),
  })
  await refreshNuxtData(AsyncDataKeys.PAPERS_PAPER(paperId))
}
